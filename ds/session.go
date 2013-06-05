package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"strconv"
	"time"
)

type session struct {
	drv             string
	db              *sql.DB
	isNumericParams bool
}

func (self *session) equalIdQuery(table *types.TableDefinition) string {
	if self.isNumericParams {
		return table.Id.Name + " = $1"
	} else {
		return table.Id.Name + " = ?"
	}
}

func replaceQuestion(buffer *bytes.Buffer, str string, idx int) (*bytes.Buffer, int) {
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
		s = s[i+1:]
		idx++
	}
	return buffer, idx
}

func (self *session) findByParams(table *types.TableDefinition,
	params map[string]string) (Query, error) {
	var buffer bytes.Buffer
	builder := &queryBuilder{table: table,
		idx:          1,
		isFirst:      true,
		buffer:       &buffer,
		operators:    default_operators,
		add_argument: (*queryBuilder).appendNumericArguments}

	if self.isNumericParams {
		builder.add_argument = (*queryBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*queryBuilder).appendSimpleArguments
	}

	e := builder.build(params)
	if nil != e {
		return nil, e
	}

	return &QueryImpl{drv: self.drv,
		db:         self.db,
		table:      table,
		where:      buffer.String(),
		parameters: builder.params}, nil
}

func (self *session) where(table *types.TableDefinition,
	queryString string, args ...interface{}) Query {

	sql := queryString
	if self.isNumericParams {
		buffer, _ := replaceQuestion(bytes.NewBuffer(make([]byte,
			0, (len(queryString)+10)*2)), queryString, 1)
		sql = buffer.String()
	}

	return &QueryImpl{drv: self.drv,
		db:         self.db,
		table:      table,
		where:      sql,
		parameters: args}
}

func (self *session) findById(table *types.TableDefinition, id interface{}) (map[string]interface{}, error) {
	query := &QueryImpl{drv: self.drv,
		db:         self.db,
		table:      table,
		where:      self.equalIdQuery(table),
		parameters: []interface{}{table.Id.Type.ToExternal(id)}}
	return query.One()
}

func (self *session) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	var buffer bytes.Buffer
	var values bytes.Buffer
	params := make([]interface{}, 0, len(table.Attributes))

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
		if self.isNumericParams {
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

	if "postgres" == self.drv {
		var id int64
		e := self.db.QueryRow(sql+" RETURNING "+table.Id.Name, params...).Scan(&id)
		if nil != e {
			return 0, e
		}
		return id, nil
	} else {
		res, e := self.db.Exec(sql, params...)
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

func (self *session) updateById(table *types.TableDefinition,
	updated_attributes map[string]interface{}, id interface{}) error {
	var buffer bytes.Buffer
	builder := &updateBuilder{table: table,
		idx:             1,
		buffer:          &buffer,
		isNumericParams: self.isNumericParams}
	e := builder.buildUpdate(updated_attributes)
	if nil != e {
		return e
	}

	if nil == builder.params || 0 == len(builder.params) {
		return errors.New("updated attributes is empty.")
	}

	builder.buildWhereById(id)

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return e
	}

	affected, e := res.RowsAffected()
	if nil != e {
		return e
	}
	if 1 != affected {
		return fmt.Errorf("affected rows is not equals 1, actual is %v", affected)
	}

	return nil
}

func (self *session) updateByParams(table *types.TableDefinition,
	updated_attributes map[string]interface{},
	params map[string]string) (int64, error) {
	var buffer bytes.Buffer
	builder := &updateBuilder{table: table,
		idx:             1,
		buffer:          &buffer,
		isNumericParams: self.isNumericParams}
	e := builder.buildUpdate(updated_attributes)
	if nil != e {
		return 0, e
	}

	if nil == builder.params || 0 == len(builder.params) {
		return 0, errors.New("updated attributes is empty.")
	}

	e = builder.buildWhere(params)
	if nil != e {
		return 0, e
	}

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *session) update(table *types.TableDefinition,
	updated_attributes map[string]interface{},
	queryString string, args ...interface{}) (int64, error) {
	var buffer bytes.Buffer
	builder := &updateBuilder{table: table,
		idx:             1,
		buffer:          &buffer,
		isNumericParams: self.isNumericParams}
	e := builder.buildUpdate(updated_attributes)
	if nil != e {
		return 0, e
	}

	if nil == builder.params || 0 == len(builder.params) {
		return 0, errors.New("updated attributes is empty.")
	}

	builder.buildWhereWithString(queryString, args)

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *session) deleteById(table *types.TableDefinition, id interface{}) error {
	var buffer bytes.Buffer
	buffer.WriteString("DELETE FROM ")
	buffer.WriteString(table.CollectionName)
	buffer.WriteString(" WHERE ")
	buffer.WriteString(table.Id.Name)
	if self.isNumericParams {
		buffer.WriteString(" = $1")
	} else {
		buffer.WriteString(" = ?")
	}

	res, e := self.db.Exec(buffer.String(), table.Id.Type.ToExternal(id))
	if nil != e {
		return e
	}

	affected, e := res.RowsAffected()
	if nil != e {
		return e
	}
	if 1 != affected {
		return fmt.Errorf("affected rows is not equals 1, actual is %v", affected)
	}

	return nil
}

func (self *session) deleteByParams(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	var buffer bytes.Buffer

	buffer.WriteString("DELETE FROM ")
	buffer.WriteString(table.CollectionName)
	builder := &queryBuilder{table: table,
		idx:       1,
		isFirst:   true,
		prefix:    " WHERE",
		buffer:    &buffer,
		operators: default_operators}

	if self.isNumericParams {
		builder.add_argument = (*queryBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*queryBuilder).appendSimpleArguments
	}

	e := builder.build(params)
	if nil != e {
		return 0, e
	}

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *session) delete(table *types.TableDefinition,
	queryString string, args ...interface{}) (int64, error) {
	var buffer bytes.Buffer

	buffer.WriteString("DELETE FROM ")
	buffer.WriteString(table.CollectionName)

	if 0 == len(queryString) {
		buffer.WriteString(" WHERE ")
		if self.isNumericParams {
			_, c := replaceQuestion(&buffer, queryString, 1)
			if len(args) != c {
				return 0, errors.New("parameters count is error")
			}
		} else {
			buffer.WriteString(queryString)
		}
	}

	res, e := self.db.Exec(buffer.String(), args...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}
