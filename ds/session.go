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

var (
	class_table_inherit_columns = []*types.ColumnDefinition{&types.ColumnDefinition{types.AttributeDefinition{Name: "tablename",
		Type:       types.GetTypeDefinition("string"),
		Collection: types.COLLECTION_UNKNOWN}},
		&types.ColumnDefinition{types.AttributeDefinition{Name: "id",
			Type:       types.GetTypeDefinition("objectId"),
			Collection: types.COLLECTION_UNKNOWN}}}

	class_table_inherit_definition = &types.TableDefinition{Name: "cti",
		UnderscoreName: "cti",
		CollectionName: "cti"}
)

func init() {

	class_table_inherit_definition.Id = class_table_inherit_columns[1]

	attributes := map[string]*types.ColumnDefinition{class_table_inherit_columns[0].Name: class_table_inherit_columns[0],
		class_table_inherit_columns[1].Name: class_table_inherit_columns[1]}

	class_table_inherit_definition.OwnAttributes = attributes
	class_table_inherit_definition.Attributes = attributes
}

type driver struct {
	drv             string
	db              *sql.DB
	isNumericParams bool
}

type session struct {
	*driver
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

////////////////////////// count //////////////////////////

func (self *session) simpleCount(table *types.TableDefinition,
	params map[string]string, isSimpleTableInheritance bool) (int64, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT count(*) FROM ")
	buffer.WriteString(table.CollectionName)

	builder := &whereBuilder{table: table,
		idx:          1,
		isFirst:      true,
		prefix:       " WHERE ",
		buffer:       &buffer,
		params:       []interface{}{},
		operators:    default_operators,
		add_argument: (*whereBuilder).appendNumericArguments}

	if self.isNumericParams {
		builder.add_argument = (*whereBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*whereBuilder).appendSimpleArguments
	}

	if isSimpleTableInheritance {
		builder.equalClass("type", table)
	}

	e := builder.build(params)
	if nil != e {
		return -1, e
	}
	row := self.db.QueryRow(buffer.String(), builder.params...)
	count := int64(0)
	e = row.Scan(&count)
	if nil != e {
		return 0, e
	}
	return count, nil
}

func (self *session) count(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	if table.IsSimpleTableInheritance() {
		return self.simpleCount(table, params, true)
	}

	if !table.HasChildren() {
		return self.simpleCount(table, params, false)
	}

	return self.simpleCount(table, params, false)
}

////////////////////////// query //////////////////////////

func (self *session) findById(table *types.TableDefinition, id string) (map[string]interface{}, error) {
	value, e := table.Id.Type.Parse(id)
	if nil != e {
		return nil, fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
			table.Id.Name, table.Id.Type.Name(), id)
	}

	builder, e := buildSQLQueryWithObjectId(self.driver, table)
	if nil != e {
		return nil, e
	}
	return builder.Bind(value).Build().One()
}

func (self *session) queryByParams(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {

	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	columns, e := buildSelectStr(table, &buffer)
	if nil != e {
		return nil, e
	}

	buffer.WriteString(" FROM ")
	buffer.WriteString(table.CollectionName)

	args, e := whereWithParams(self.driver, table, false, 1, params, &buffer)
	if nil != e {
		return nil, e
	}

	q := &QueryImpl{drv: self.driver, table: table, columns: columns, sql: buffer.String(), parameters: args}

	return q.All()
}

func (self *session) queryWithParamsBySimpleTableInheritance(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT * FROM ")
	buffer.WriteString(table.CollectionName)
	args, e := whereWithParams(self.driver, table, true, 1, params, &buffer)
	if nil != e {
		return nil, e
	}
	q := &QueryImpl{drv: self.driver,
		isSimpleTableInheritance: true,
		table:                    table,
		sql:                      buffer.String(),
		parameters:               args}

	return q.All()
}

func (self *session) queryWithParamsByClassTableInheritance(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {

	var buffer bytes.Buffer
	buffer.WriteString("SELECT tableoid::regclass as tablename, id as id FROM ")
	buffer.WriteString(table.CollectionName)
	args, e := whereWithParams(self.driver, table, false, 1, params, &buffer)
	if nil != e {
		return nil, e
	}
	q := &QueryImpl{drv: self.driver,
		table:      class_table_inherit_definition,
		columns:    class_table_inherit_columns,
		sql:        buffer.String(),
		parameters: args}

	id_list, e := q.All()
	if nil != e {
		return nil, e
	}

	var last_builder QueryBuilder
	var last_name string

	results := make([]map[string]interface{}, 0, len(id_list))
	for _, instance_id := range id_list {
		tablename, ok := instance_id["tablename"]
		if !ok {
			panic("'tablename' is not found")
		}

		name, ok := tablename.(string)
		if !ok {
			panic("'tablename' is not a string")
		}

		id, ok := instance_id["id"]
		if !ok {
			panic("'tablename' is not found")
		}
		if last_name != name {
			if nil == table.Children {
				if name != table.CollectionName {
					return nil, errors.New("table '" + name + "' is undefined.")
				}
				last_builder, e = buildSQLQueryWithObjectId(self.driver, table)
			} else {
				realTable := table.Children.FindByTableName(name)
				if nil == realTable {
					return nil, errors.New("table '" + name + "' is undefined.")
				}
				last_builder, e = buildSQLQueryWithObjectId(self.driver, table)
			}
			if nil != e {
				return nil, e
			}
			last_name = name
		}

		if nil == last_builder {
			return nil, errors.New("table '" + name + "' is undefined.")
		}

		instance, e := last_builder.Bind(id).Build().One()
		if nil != e {
			return nil, e
		}

		results = append(results, instance)
	}
	return results, nil
}

func (self *session) query(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {

	if table.IsSimpleTableInheritance() {
		return self.queryWithParamsBySimpleTableInheritance(table, params)
	}

	if !table.HasChildren() {
		return self.queryByParams(table, params)
	}

	return self.queryWithParamsByClassTableInheritance(table, params)
}

////////////////////////// insert //////////////////////////

func (self *session) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	if t, ok := attributes["type"]; ok {
		nm, ok := t.(string)
		if !ok {
			return 0, fmt.Errorf("'type' must is a string, actual type is a %T, actual value is '%v'.", t, t)
		}
		if !table.HasChildren() {
			return 0, errors.New("table '" + nm + "' is not exists.")
		}
		defintion := table.Children.FindByUnderscoreName(nm)
		if nil == defintion {
			return 0, errors.New("table '" + nm + "' is not exists.")
		}

		if !defintion.IsSubclassOf(table) {
			return 0, errors.New("table '" + nm + "' is not inherit from table '" + table.UnderscoreName + "'")
		}

		table = defintion
	}

	var buffer bytes.Buffer
	var values bytes.Buffer
	params := make([]interface{}, 0, len(table.Attributes))

	idx := 1
	var e error
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
		case "type":
			value = table.UnderscoreName
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

			value, e = attribute.Type.ToInternal(v)
			if nil != e {
				return 0, fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
					attribute.Name, attribute.Type.Name(), v)
			}
			value = attribute.Type.ToExternal(value)
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
		e = self.db.QueryRow(sql+" RETURNING "+table.Id.Name, params...).Scan(&id)
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

////////////////////////// update //////////////////////////

func (self *session) update(table *types.TableDefinition,
	params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {
	if table.IsSimpleTableInheritance() {
		return self.updateByParams(table, true, params, updated_attributes)
	}

	if !table.HasChildren() {
		return self.updateByParams(table, false, params, updated_attributes)
	}

	return self.updateWithParamsByClassTableInheritance(table, params, updated_attributes)
}

func (self *session) updateById(table *types.TableDefinition, id string,
	updated_attributes map[string]interface{}) error {
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
	value, e := table.Id.Type.Parse(id)
	if nil != e {
		return fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
			table.Id.Name, table.Id.Type.Name(), id)
	}

	builder.buildWhereById(value)

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

func (self *session) updateWithParamsByClassTableInheritance(table *types.TableDefinition,
	params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {

	effected_all := int64(0)
	for _, child := range table.OwnChildren.All() {
		effected_single, e := self.update(child, params, updated_attributes)
		if nil != e {
			return 0, e
		}
		effected_all += effected_single
	}
	return effected_all, nil
}

func (self *session) updateByParams(table *types.TableDefinition,
	isSimpleTableInheritance bool,
	params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {
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

	e = builder.buildWhere(params, isSimpleTableInheritance)
	if nil != e {
		return 0, e
	}

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *session) updateBySQL(table *types.TableDefinition,
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

////////////////////////// update //////////////////////////

func (self *session) delete(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	if table.IsSimpleTableInheritance() {
		return self.deleteByParams(table, true, params)
	}

	if !table.HasChildren() {
		return self.deleteByParams(table, false, params)
	}

	return self.deleteWithParamsByClassTableInheritance(table, params)
}

func (self *session) deleteById(table *types.TableDefinition, id string) error {
	value, e := table.Id.Type.Parse(id)
	if nil != e {
		return fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
			table.Id.Name, table.Id.Type.Name(), id)
	}

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

	res, e := self.db.Exec(buffer.String(), table.Id.Type.ToExternal(value))
	if nil != e {
		return e
	}

	affected, e := res.RowsAffected()
	if nil != e {
		return e
	}
	if 1 != affected {
		return sql.ErrNoRows
	}

	return nil
}

func (self *session) deleteByParams(table *types.TableDefinition,
	isSimpleTableInheritance bool,
	params map[string]string) (int64, error) {
	var buffer bytes.Buffer

	buffer.WriteString("DELETE FROM ")
	buffer.WriteString(table.CollectionName)
	builder := &whereBuilder{table: table,
		idx:       1,
		isFirst:   true,
		prefix:    " WHERE ",
		buffer:    &buffer,
		operators: default_operators}

	if self.isNumericParams {
		builder.add_argument = (*whereBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*whereBuilder).appendSimpleArguments
	}

	if isSimpleTableInheritance {
		builder.equalClass("type", table)
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

func (self *session) deleteWithParamsByClassTableInheritance(table *types.TableDefinition,
	params map[string]string) (int64, error) {

	effected_all := int64(0)
	for _, child := range table.OwnChildren.All() {
		effected_single, e := self.delete(child, params)
		if nil != e {
			return 0, e
		}
		effected_all += effected_single
	}
	return effected_all, nil
}

func (self *session) deleteBySQL(table *types.TableDefinition,
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
