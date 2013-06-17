package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

const (
	GENEERIC_DB = 0
	POSTGRESQL  = 1
	MSSQL       = 2
	ORACLE      = 3
	SQLITE      = 4
	MYSQL       = 5
)

func GetDBType(drv string) int {
	switch drv {
	case "postgres":
		return POSTGRESQL
	case "mssql":
		return MSSQL
	case "sqlite":
		return SQLITE
	case "oracle":
		return ORACLE
	case "mysql":
		return MYSQL
	default:
		return GENEERIC_DB
	}
}

var (
	id_column        *types.ColumnDefinition = nil
	tablename_column *types.ColumnDefinition = nil

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
	tablename_column = class_table_inherit_columns[0]
	id_column = class_table_inherit_columns[1]
	class_table_inherit_definition.Id = class_table_inherit_columns[1]

	attributes := map[string]*types.ColumnDefinition{class_table_inherit_columns[0].Name: class_table_inherit_columns[0],
		class_table_inherit_columns[1].Name: class_table_inherit_columns[1]}

	class_table_inherit_definition.OwnAttributes = attributes
	class_table_inherit_definition.Attributes = attributes
}

type driver struct {
	drv             string
	dbType          int
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

func (self *session) findById(table *types.TableDefinition, id, includes string) (map[string]interface{}, error) {
	value, e := table.Id.Type.Parse(id)
	if nil != e {
		return nil, fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
			table.Id.Name, table.Id.Type.Name(), id)
	}

	if table.IsSingleTableInheritance() {
		goto default_query
	}

	if !table.HasChildren() {
		goto default_query
	}

	return self.findByIdAndClassTableInheritance(table, id, includes)

default_query:
	builder, e := buildSQLQueryWithObjectId(self.driver, table)
	if nil != e {
		return nil, e
	}
	result, e := builder.Bind(value).Build().One()

	if nil != e {
		return nil, e
	}

	if 0 != len(includes) {
		e = self.loadIncludes(table, result, includes)
		if nil != e {
			return nil, e
		}
	}
	return result, nil
}

func (self *session) findByIdAndClassTableInheritance(table *types.TableDefinition, id, includes string) (map[string]interface{}, error) {
	value, e := table.Id.Type.Parse(id)
	if nil != e {
		return nil, fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
			table.Id.Name, table.Id.Type.Name(), id)
	}

	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	columns := toColumns(table, false)
	buffer.WriteString("tableoid::regclass as tablename, ")
	if nil == columns || 0 == len(columns) {
		return nil, errors.New("crazy! selected columns is empty.")
	}
	writeColumns(columns, &buffer)
	buffer.WriteString(" FROM ")
	buffer.WriteString(table.CollectionName)
	buffer.WriteString(" WHERE ")
	buffer.WriteString(table.Id.Name)
	if self.isNumericParams {
		buffer.WriteString(" = $1")
	} else {
		buffer.WriteString(" = ?")
	}

	new_columns := make([]*types.ColumnDefinition, len(columns)+1)
	new_columns[0] = tablename_column
	copy(new_columns[1:], columns)

	values, e := selectOne(self.driver, buffer.String(), []interface{}{value}, new_columns...)
	if nil != e {
		return nil, e
	}

	tablename, ok := values[0].(string)
	if !ok {
		return nil, errors.New("table name is not a string")
	}
	if tablename != table.CollectionName {
		new_table := table.FindByTableName(tablename)
		if nil == new_table {
			return nil, errors.New("table name '" + tablename + "' is undefined.")
		}
		return self.findById(new_table, id, includes)
	}
	res := make(map[string]interface{})
	res["type"] = table.UnderscoreName
	for i := 1; i < len(new_columns); i++ {
		res[new_columns[i].Name] = values[i]
	}

	if 0 != len(includes) {
		e = self.loadIncludes(table, res, includes)
		if nil != e {
			return nil, e
		}
	}
	return res, nil
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
	if self.driver.dbType == POSTGRESQL {
		buffer.WriteString(" ONLY ")
	}
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

	var results []map[string]interface{}
	var e error

	if table.IsSingleTableInheritance() {
		results, e = self.queryByParams(table, true, params)
		goto end
	}

	if !table.HasChildren() {
		results, e = self.queryByParams(table, false, params)
		goto end
	}

	results, e = self.queryByParamsAndClassTableInheritance(table, params)
end:
	if nil != e {
		return nil, e
	}

	if includes, ok := params["includes"]; ok && 0 != len(includes) {
		for _, result := range results {
			e = self.loadIncludes(table, result, includes)
			if nil != e {
				return nil, e
			}
		}
	}
	return results, e
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

func (self *session) loadIncludes(parent_table *types.TableDefinition, parent map[string]interface{}, includes string) error {
	parent_id := parent["id"]
	if nil == parent_id {
		return errors.New("parent id is nil while load children.")
	}
	parent_id_str := fmt.Sprint(parent_id)

	if "*" == includes {
		assocations := parent_table.GetAssocationByTypes(types.HAS_ONE, types.HAS_MANY)

		if nil == assocations || 0 == len(assocations) {
			return nil
		}
		for _, assocation := range assocations {
			results, e := self.findByParent(parent_table, parent_id_str, assocation, assocation.Target())
			if nil != e {
				return e
			}
			parent["$"+assocation.Target().UnderscoreName] = results
		}
	} else {
		for _, s := range strings.Split(includes, ",") {
			target := self.tables.FindByUnderscoreName(s)
			if nil == target {
				return errors.New("table '" + s + "' is not found in the includes.")
			}
			assocations := parent_table.GetAssocationByTargetAndTypes(target, types.HAS_ONE, types.HAS_MANY)
			if nil == assocations || 0 == len(assocations) {
				return errors.New("assocation that to '" + s + "' is not found in the includes.")
			}

			for _, assocation := range assocations {
				results, e := self.findByParent(parent_table, parent_id_str, assocation, target)
				if nil != e {
					return e
				}
				parent["$"+target.UnderscoreName] = results
			}

		}
	}
	return nil
}

func (self *session) findByParent(parent_table *types.TableDefinition,
	parent_id string, assocation types.Assocation,
	target *types.TableDefinition) ([]map[string]interface{}, error) {
	var foreignKey string
	var is_polymorphic bool
	switch assocation.Type() {
	case types.HAS_ONE:
		hasOne := assocation.(*types.HasOne)
		is_polymorphic = hasOne.Polymorphic
		foreignKey = hasOne.ForeignKey
	case types.HAS_MANY:
		hasMany := assocation.(*types.HasMany)
		is_polymorphic = hasMany.Polymorphic
		foreignKey = hasMany.ForeignKey
	default:
		return nil, errors.New("unsupported assocation type - " + assocation.Type().String())
	}

	params := map[string]string{}
	if is_polymorphic {
		params["@parent_type"] = parent_table.UnderscoreName
		params["@parent_id"] = parent_id
	} else {
		params["@"+foreignKey] = parent_id
	}
	return self.query(target, params)
}

func (self *session) children(parent_table *types.TableDefinition, parent_id string,
	target *types.TableDefinition, foreignKey string) ([]map[string]interface{}, error) {
	assocation, e := parent_table.GetAssocation(target, foreignKey, types.HAS_MANY, types.HAS_ONE)
	if nil != e {
		return nil, e
	}
	return self.findByParent(parent_table, parent_id, assocation, target)
}

func (self *session) parentBy(child_table *types.TableDefinition, child_id string,
	assocation types.Assocation, target *types.TableDefinition) (map[string]interface{}, error) {
	var foreignKey string
	var is_polymorphic bool
	switch assocation.Type() {
	case types.HAS_ONE:
		hasOne := assocation.(*types.HasOne)
		is_polymorphic = hasOne.Polymorphic
		foreignKey = hasOne.ForeignKey
	case types.HAS_MANY:
		hasMany := assocation.(*types.HasMany)
		is_polymorphic = hasMany.Polymorphic
		foreignKey = hasMany.ForeignKey
	default:
		return nil, errors.New("unsupported assocation type - " + assocation.Type().String())
	}

	res, e := self.findById(child_table, child_id, "")
	if nil != e {
		return nil, e
	}

	if is_polymorphic {
		v := res["parent_type"]
		if nil == v {
			return nil, errors.New("'parent_type' is nil in the result")
		}

		parent_type, ok := v.(string)
		if !ok {
			return nil, errors.New("'parent_type' is not a string in the result")
		}

		parent := self.tables.FindByUnderscoreName(parent_type)
		if nil == parent {
			return nil, errors.New(" table '" + parent_type + "' is not exists.")
		}

		if parent != target && parent.IsSubclassOf(target) {
			return nil, errors.New(" table '" + parent_type +
				"' is not a subclass of table '" + target.UnderscoreName + "'.")
		}

		v = res["parent_id"]
		if nil == v {
			return nil, errors.New("'parent_id' is nil in the result")
		}

		parent_id := fmt.Sprint(v)
		return self.findById(parent, parent_id, "")
	}

	id := res[foreignKey]
	if nil == id {
		return nil, errors.New("'" + foreignKey + "' is not exists in the result.")
	}

	return self.findById(target, fmt.Sprint(id), "")

}

func (self *session) parent(child_table *types.TableDefinition, child_id string,
	target *types.TableDefinition, foreignKey string) (map[string]interface{}, error) {
	assocation, e := child_table.GetAssocation(target, foreignKey, types.BELONGS_TO)
	if nil != e {
		assocation, e := target.GetAssocation(child_table, foreignKey, types.HAS_MANY, types.HAS_ONE)
		if nil != e {
			return nil, e
		}
		return self.parentBy(child_table, child_id, assocation, target)
	}
	belongsTo := assocation.(*types.BelongsTo)
	res, e := self.findById(child_table, child_id, "")
	if nil != e {
		return nil, e
	}
	id := res[belongsTo.Name.Name]
	if nil == id {
		return nil, errors.New("'" + belongsTo.Name.Name + "' is not exists in the result.")
	}

	return self.findById(target, fmt.Sprint(id), "")
}

////////////////////////// insert //////////////////////////
func typeFrom(table *types.TableDefinition, attributes map[string]interface{}) (*types.TableDefinition, error) {
	t, ok := attributes["type"]
	if !ok {
		return table, nil
	}

	nm, ok := t.(string)
	if !ok {
		return nil, fmt.Errorf("'type' must is a string, actual type is a %T, actual value is '%v'.", t, t)
	}

	if nm == table.UnderscoreName {
		return table, nil
	}

	defintion := table.FindByUnderscoreName(nm)
	if nil == defintion {
		return nil, errors.New("table '" + nm + "' is not exists.")
	}

	if !defintion.IsSubclassOf(table) {
		return nil, errors.New("table '" + nm + "' is not inherit from table '" + table.UnderscoreName + "'")
	}

	return defintion, nil
}

func (self *session) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	var e error
	table, e = typeFrom(table, attributes)
	if nil != e {
		return 0, e
	}

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

	var id int64
	if "postgres" == self.drv {
		e = self.db.QueryRow(sql+" RETURNING "+table.Id.Name, params...).Scan(&id)
		if nil != e {
			return 0, e
		}
	} else {
		res, e := self.db.Exec(sql, params...)
		if nil != e {
			return 0, e
		}

		if affected, e := res.RowsAffected(); 1 != affected {
			return 0, fmt.Errorf("insert to %v failed, affected rows is %v %v",
				table.CollectionName, affected, e)
		}

		id, e = res.LastInsertId()
		if nil != e {
			return 0, e
		}
	}

	e = self.createChildren(table, id, attributes)
	if nil != e {
		return 0, e
	}
	return id, nil
}

func (self *session) createChildren(parent_table *types.TableDefinition,
	parent_id interface{}, attributes map[string]interface{}) error {
	for name, v := range attributes {
		if '$' != name[0] {
			continue
		}

		target_table := self.tables.FindByUnderscoreName(name[1:])
		if nil == target_table {
			return fmt.Errorf("table '%s' with '%s' is not found ", name[1:], name)
		}

		assoc := parent_table.GetAssocationByTargetAndTypes(target_table, types.HAS_MANY, types.HAS_MANY)
		if nil == assoc {
			return fmt.Errorf("table '%s' is not contains child that name is '%s' at the '%s'",
				parent_table.Name, name[1:], name)
		}
		if 1 != len(assoc) {
			return fmt.Errorf("table '%s' is contains %v children that name is '%s' at the '%s'",
				parent_table.Name, len(assoc), name[1:], name)
		}

		switch a := assoc[0].(type) {
		case *types.HasMany:
			if values, ok := v.([]map[string]interface{}); ok {
				for _, value := range values {
					_, e := self.createChild(target_table, parent_table, parent_id, value, a.ForeignKey, a.Polymorphic)
					if nil != e {
						return errors.New("save attributes to '" + target_table.Name + "' failed, " + e.Error())
					}
				}
			} else if values, ok := v.([]interface{}); ok {
				for _, value := range values {
					attrs, ok := value.(map[string]interface{})
					if !ok {
						return fmt.Errorf("value of '%s' is not map[string]interface{}", name)
					}

					_, e := self.createChild(target_table, parent_table, parent_id, attrs, a.ForeignKey, a.Polymorphic)
					if nil != e {
						return errors.New("save attributes to '" + target_table.Name + "' failed, " + e.Error())
					}
				}
			} else {
				return fmt.Errorf("value of '%s' is not []map[string]interface{}, actual is %T", name, v)
			}
		case *types.HasOne:
			attrs, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("value of '%s' is not map[string]interface{}", name)
			}
			_, e := self.createChild(target_table, parent_table, parent_id, attrs, a.ForeignKey, a.Polymorphic)
			if nil != e {
				return errors.New("save attributes to '" + target_table.Name + "' failed, " + e.Error())
			}
		default:
			panic("between '" + parent_table.Name + "' and '" + target_table.Name + "' is not hasMany or hasOne")
		}
	}
	return nil
}

func (self *session) createChild(table, parent_table *types.TableDefinition, parent_id interface{},
	attributes map[string]interface{}, foreignKey string, is_polymorphic bool) (int64, error) {
	var e error
	table, e = typeFrom(table, attributes)
	if nil != e {
		return 0, e
	}

	if is_polymorphic {
		attributes["parent_type"] = parent_table.UnderscoreName
		attributes["parent_id"] = parent_id
	} else {
		attributes[foreignKey] = parent_id
	}

	return self.insert(table, attributes)
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
		dbType:          self.driver.dbType,
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
		dbType:          self.driver.dbType,
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
		dbType:          self.driver.dbType,
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
	if self.driver.dbType == POSTGRESQL {
		buffer.WriteString(" ONLY ")
	}
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
	if self.driver.dbType == POSTGRESQL {
		buffer.WriteString(" ONLY ")
	}
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
	if self.driver.dbType == POSTGRESQL {
		buffer.WriteString(" ONLY ")
	}
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
