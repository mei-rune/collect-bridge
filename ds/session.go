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
	id_column *types.ColumnDefinition = nil

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
	id_column = class_table_inherit_columns[1]
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
	tables *types.TableDefinitions
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

func (self *session) newWhere(idx int,
	table *types.TableDefinition,
	buffer *bytes.Buffer) *whereBuilder {

	builder := &whereBuilder{tables: self.tables,
		table:     table,
		idx:       idx,
		isFirst:   true,
		prefix:    " WHERE ",
		buffer:    buffer,
		operators: default_operators,
		operators_for_field: map[string]map[string]op_func{"type": operators_for_type,
			"parent_type": operators_for_parent_type},
		add_argument: (*whereBuilder).appendNumericArguments}

	if self.isNumericParams {
		builder.add_argument = (*whereBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*whereBuilder).appendSimpleArguments
	}

	return builder
}

////////////////////////// count //////////////////////////

func (self *session) simpleCount(table *types.TableDefinition,
	params map[string]string, isSingleTableInheritance bool) (int64, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT count(*) FROM ")
	buffer.WriteString(table.CollectionName)

	builder := self.newWhere(1, table, &buffer)

	if isSingleTableInheritance {
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
	if table.IsSingleTableInheritance() {
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

	if table.IsSingleTableInheritance() {
		goto default_do
	}

	if !table.HasChildren() {
		goto default_do
	}

	return self.findByIdAndClassTableInheritance(table, id)

default_do:
	builder, e := buildSQLQueryWithObjectId(self.driver, table)
	if nil != e {
		return nil, e
	}
	return builder.Bind(value).Build().One()
}

func (self *session) findByIdAndClassTableInheritance(table *types.TableDefinition, id string) (map[string]interface{}, error) {
	return nil, errors.New("findByIdAndClassTableInheritance is not implmented")
}

func (self *session) queryByParams(table *types.TableDefinition,
	isSingleTableInheritance bool,
	params map[string]string) ([]map[string]interface{}, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")

	columns := toColumns(table, isSingleTableInheritance)
	if nil == columns || 0 == len(columns) {
		return nil, errors.New("crazy! selected columns is empty.")
	}
	writeColumns(columns, &buffer)

	buffer.WriteString(" FROM ")
	buffer.WriteString(table.CollectionName)
	args, e := self.whereWithParams(table, isSingleTableInheritance, 1, params, &buffer)
	if nil != e {
		return nil, e
	}

	q := &QueryImpl{drv: self.driver,
		isSingleTableInheritance: isSingleTableInheritance,
		columns:                  columns,
		table:                    table,
		sql:                      buffer.String(),
		parameters:               args}

	return q.All()
}

func firstTable(tables *types.TableDefinitions) *types.TableDefinition {
	for _, t := range tables.All() {
		return t
	}
	return nil
}

func (self *session) queryByParamsAndClassTableInheritance(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {

	var buffer bytes.Buffer
	buffer.WriteString("SELECT tableoid::regclass as tablename, id as id FROM ")
	buffer.WriteString(table.CollectionName)
	args, e := self.whereWithParams(table, false, 1, params, &buffer)
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
			table := table.FindByTableName(name)
			if nil == table {
				return nil, errors.New("table '" + name + "' is undefined.")
			}

			last_builder, e = buildSQLQueryWithObjectId(self.driver, table)
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

	if table.IsSingleTableInheritance() {
		return self.queryByParams(table, true, params)
	}

	if !table.HasChildren() {
		return self.queryByParams(table, false, params)
	}

	return self.queryByParamsAndClassTableInheritance(table, params)
}

func (self *session) whereWithParams(table *types.TableDefinition, isSingleTableInheritance bool,
	idx int, params map[string]string, buffer *bytes.Buffer) ([]interface{}, error) {
	if nil == params {
		if !isSingleTableInheritance {
			return nil, nil
		}
	}

	builder := self.newWhere(idx, table, buffer)

	if isSingleTableInheritance {
		builder.equalClass("type", table)
	}

	if nil == params {
		return nil, nil
	}

	e := builder.build(params)
	if nil != e {
		return nil, e
	}

	if groupBy, ok := params["groupBy"]; ok {
		if 0 != len(groupBy) {
			return nil, errors.New("groupBy is empty.")
		}

		buffer.WriteString(" GROUP BY ")
		buffer.WriteString(groupBy)
	}

	if having, ok := params["having"]; ok {
		if 0 != len(having) {
			return nil, errors.New("having is empty.")
		}

		buffer.WriteString(" HAVING ")
		buffer.WriteString(having)
	}

	if order, ok := params["order"]; ok {
		if 0 != len(order) {
			return nil, errors.New("order is empty.")
		}

		buffer.WriteString(" ORDER BY ")
		buffer.WriteString(order)
	}

	if limit, ok := params["limit"]; ok {
		i, e := strconv.ParseInt(limit, 10, 64)
		if nil != e {
			return nil, fmt.Errorf("limit is not a number, actual value is '" + limit + "'")
		}
		if i <= 0 {
			return nil, fmt.Errorf("limit must is geater zero, actual value is '" + limit + "'")
		}

		if offset, ok := params["offset"]; ok {
			i, e = strconv.ParseInt(offset, 10, 64)
			if nil != e {
				return nil, fmt.Errorf("offset is not a number, actual value is '" + offset + "'")
			}

			if i < 0 {
				return nil, fmt.Errorf("offset must is geater(or equals) zero, actual value is '" + offset + "'")
			}

			buffer.WriteString(" LIMIT ")
			buffer.WriteString(offset)
			buffer.WriteString(" , ")
			buffer.WriteString(limit)
		} else {
			buffer.WriteString(" LIMIT ")
			buffer.WriteString(limit)
		}
	}

	return builder.params, nil
}

////////////////////////// insert //////////////////////////

func (self *session) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	if t, ok := attributes["type"]; ok {
		nm, ok := t.(string)
		if !ok {
			return 0, fmt.Errorf("'type' must is a string, actual type is a %T, actual value is '%v'.", t, t)
		}

		if nm != table.UnderscoreName {
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
	if table.IsSingleTableInheritance() {
		return self.updateByParams(table, true, params, updated_attributes)
	}

	if !table.HasChildren() {
		return self.updateByParams(table, false, params, updated_attributes)
	}

	return self.updateByParamsAndClassTableInheritance(table, params, updated_attributes)
}

func (self *session) updateByParamsAndClassTableInheritance(table *types.TableDefinition,
	params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {

	effected_single, e := self.updateByParams(table, false, params, updated_attributes)
	if nil != e {
		return 0, e
	}
	effected_all := effected_single

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
	isSingleTableInheritance bool,
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

	e = builder.buildWhere(params, isSingleTableInheritance)
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

////////////////////////// update //////////////////////////

func (self *session) delete(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	if table.IsSingleTableInheritance() {
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

	_, e = self.deleteCascadeById(table, id)
	if nil != e {
		return e
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
	isSingleTableInheritance bool,
	params map[string]string) (int64, error) {

	_, e := self.deleteCascadeByParams(table, isSingleTableInheritance, params)
	if nil != e {
		return 0, e
	}

	var buffer bytes.Buffer
	buffer.WriteString("DELETE FROM ")
	buffer.WriteString(table.CollectionName)

	builder := self.newWhere(1, table, &buffer)
	if isSingleTableInheritance {
		builder.equalClass("type", table)
	}

	e = builder.build(params)
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

	effected_single, e := self.deleteByParams(table, false, params)
	if nil != e {
		return 0, e
	}
	effected_all := effected_single

	for _, child := range table.OwnChildren.All() {
		effected_single, e = self.delete(child, params)
		if nil != e {
			return 0, e
		}
		effected_all += effected_single
	}
	return effected_all, nil
}

func (self *session) deleteBySQL(table *types.TableDefinition,
	queryString string, args ...interface{}) (int64, error) {

	_, e := self.deleteCascadeBySQL(table, queryString, args...)
	if nil != e {
		return 0, e
	}

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
