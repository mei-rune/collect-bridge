package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

func replaceQuestion(buffer *bytes.Buffer, str string, idx int) *bytes.Buffer {
	s := []byte(str)
	for {
		i := bytes.IndexByte(s, '?')
		if -1 == i {
			buffer.Write(s)
			break
		}
		buffer.Write(s[:i])
		buffer.WriteString("$")
		buffer.WriteString(strconv.FormatInt(int64(idx), 10))

		// if (idx + 1) == len(s) {
		// 	break
		// }
		s = s[i+1:]

		idx++
	}
	return buffer
}

func FindByParams(drv string, db *sql.DB, table *types.TableDefinition,
	params map[string]string) (Query, error) {
	var buffer bytes.Buffer
	isNumParams := "pg" == drv
	builder := &statementBuilder{table: table,
		idx:          1,
		isFirst:      true,
		buffer:       &buffer,
		operators:    default_operators,
		add_argument: (*statementBuilder).appendNumericArguments}

	if isNumParams {
		builder.add_argument = (*statementBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*statementBuilder).appendSimpleArguments
	}

	e := builder.build(params)
	if nil != e {
		return nil, e
	}

	//fmt.Printf("%v, %v\r\n", buffer.String(), builder.params)

	return &QueryImpl{drv: drv,
		db:         db,
		table:      table,
		where:      buffer.String(),
		parameters: builder.params}, nil
}

func Where(drv string, db *sql.DB, table *types.TableDefinition,
	queryString string, args ...interface{}) Query {

	isNumParams := "pg" == drv
	sql := queryString
	if isNumParams {
		sql = replaceQuestion(bytes.NewBuffer(make([]byte,
			0, (len(queryString)+10)*2)), queryString, 1).String()
	}

	return &QueryImpl{drv: drv,
		db:         db,
		table:      table,
		where:      sql,
		parameters: args}
}

func FindById(drv string, db *sql.DB, table *types.TableDefinition,
	id interface{}) (map[string]interface{}, error) {
	isNumParams := "pg" == drv

	queryStr := table.Id.Name + " = ?"
	if isNumParams {
		queryStr = table.Id.Name + " = $1"
	}

	query := &QueryImpl{drv: drv,
		db:         db,
		table:      table,
		where:      queryStr,
		parameters: []interface{}{table.Id.Type.ToExternal(id)}}
	return query.One()
}

func Insert(drv string, db *sql.DB, table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	var buffer bytes.Buffer
	var values bytes.Buffer
	params := make([]interface{}, 0, len(table.Attributes))

	isNumParams := "pg" == drv

	idx := 1
	for _, attribute := range table.Attributes {
		//////////////////////////////////////////
		// TODO: refactor it?
		if attribute.IsSerial() {
			continue
		}

		var value interface{} = nil
		switch attribute.Name {
		case "created_at":
			fallthrough
		case "updated_at":
			value = time.Now()
		default:
			v := attributes[attribute.Name]
			if nil == v {
				if attribute.IsRequired {
					return 0, fmt.Errorf("column '%v' is required", attribute.Name)
				}
				if nil == attribute.DefaultValue {
					continue
				}
				v = attribute.DefaultValue
			}
			value = attribute.Type.ToExternal(v)
		}
		//////////////////////////////////////////

		if 1 != idx {
			buffer.WriteString(", ")
			values.WriteString(", ")
		}
		buffer.WriteString(attribute.Name)
		if isNumParams {
			values.WriteString("$")
			values.WriteString(strconv.FormatInt(int64(idx), 10))
		} else {
			values.WriteString("?")
		}
		params = append(params, value)

		idx++
	}

	sql := "INSERT INTO " + table.CollectionName + "( " + buffer.String() +
		" ) VALUES ( " + values.String() + " )"

	if "pg" == drv {
		var id int64
		e := db.QueryRow(sql+" RETURNING "+table.Id.Name, params...).Scan(&id)
		if nil != e {
			return 0, e
		}
		return id, nil
	} else {
		res, e := db.Exec(sql, params...)
		if nil != e {
			return 0, e
		}

		if affected, e := res.RowsAffected(); 1 != affected {
			return 0, fmt.Errorf("insert to %v failed, affected rows is %v %v",
				table.CollectionName, affected, e)
		}

		return res.LastInsertId()
	}
}

func UpdateById(drv string, db *sql.DB, table *types.TableDefinition,
	updated_attributes map[string]interface{}, id interface{}) error {
	_, e := Update(drv, db, table, updated_attributes, table.Id.Name+" = ?", id)
	return e
}

func Update(drv string, db *sql.DB, table *types.TableDefinition,
	updated_attributes map[string]interface{}, queryString string, args ...interface{}) (int64, error) {
	var buffer bytes.Buffer
	var params []interface{} = make([]interface{}, 0, len(table.Attributes))

	buffer.WriteString("UPDATE ")
	buffer.WriteString(table.CollectionName)
	buffer.WriteString(" SET ")

	isNumParams := "pg" == drv

	idx := 1
	for _, attribute := range table.Attributes {
		//////////////////////////////////////////
		// TODO: refactor it?
		if attribute.IsSerial() {
			continue
		}
		var value interface{} = nil
		switch attribute.Name {
		case "created_at":
			continue
		case "updated_at":
			value = time.Now()
		default:
			v := updated_attributes[attribute.Name]
			if nil == v {
				continue
			}
			if attribute.IsReadOnly {
				return 0, fmt.Errorf("column '%v' is readonly.", attribute.Name)
			}
			value = attribute.Type.ToExternal(v)
		}

		//////////////////////////////////////////

		if 1 != idx {
			buffer.WriteString(", ")
		}

		buffer.WriteString(attribute.Name)
		if isNumParams {
			buffer.WriteString("= $")
			buffer.WriteString(strconv.FormatInt(int64(idx), 10))
		} else {
			buffer.WriteString(" = ?")
		}

		params = append(params, value)

		idx++
	}
	buffer.WriteString(" WHERE ")

	if isNumParams {
		replaceQuestion(&buffer, queryString, idx)
	} else {
		buffer.WriteString(queryString)
	}

	// for nm, v := range args {
	// 	// attribute := table.GetAttribute(nm)
	// 	// if nil == attribute {
	// 	// 	return 0, fmt.Errorf("column '%v' is not exists.", nm)
	// 	// }
	// 	params = append(params, attribute.Type.ToExternal(v))
	// }
	params = append(params, args...)

	res, e := db.Exec(buffer.String(), params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func DeleteById(drv string, db *sql.DB, table *types.TableDefinition,
	id interface{}) error {
	_, e := Delete(drv, db, table, table.Id.Name+" = ?", id)
	return e
}

func Delete(drv string, db *sql.DB, table *types.TableDefinition,
	queryString string, args ...interface{}) (int64, error) {
	var buffer bytes.Buffer
	var params []interface{} = make([]interface{}, 0, len(table.Attributes))

	buffer.WriteString("DELETE FROM ")
	buffer.WriteString(table.CollectionName)

	if 0 == len(queryString) {

		isNumParams := "pg" == drv
		buffer.WriteString(" WHERE ")
		if isNumParams {
			replaceQuestion(&buffer, queryString, 1)
		} else {
			buffer.WriteString(queryString)
		}

		// for nm, v := range args {
		// 	// attribute := table.GetAttribute(nm)
		// 	// if nil == attribute {
		// 	// 	return 0, fmt.Errorf("column '%v' is not exists.", nm)
		// 	// }
		// 	params = append(params, attribute.Type.ToExternal(v))
		// }
		params = append(params, args...)
	}

	res, e := db.Exec(buffer.String(), params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}
