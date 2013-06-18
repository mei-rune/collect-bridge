package ds

import (
	"commons/types"
	"database/sql"
)

type default_cti_policy struct {
	simple driver
	cti    driver
}

func (self *default_cti_policy) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	return self.simple.insert(table, attributes)
}

func (self *default_cti_policy) count(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	effected_single, e := self.simple.count(table, params)
	if nil != e {
		return 0, e
	}
	effected_all := effected_single

	for _, child := range table.OwnChildren.All() {
		effected_single, e := self.cti.count(child, params)
		if nil != e {
			return 0, e
		}
		effected_all += effected_single
	}
	return effected_all, nil
}

func (self *default_cti_policy) findById(table *types.TableDefinition,
	id interface{}) (map[string]interface{}, error) {
	result, e := self.simple.findById(table, id)
	if nil == e {
		return result, nil
	}
	if e != sql.ErrNoRows {
		return nil, e
	}

	for _, child := range table.OwnChildren.All() {
		result, e = self.cti.findById(child, id)
		if nil == e {
			return result, nil
		}

		if e != sql.ErrNoRows {
			return nil, e
		}
	}

	return nil, sql.ErrNoRows
}

func (self *default_cti_policy) find(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {
	results, e := self.simple.find(table, params)
	if nil != e && e != sql.ErrNoRows {
		return nil, e
	}

	for _, child := range table.OwnChildren.All() {
		results_single, e := self.cti.find(child, params)
		if nil != e && e != sql.ErrNoRows {
			return nil, e
		}
		results = append(results, results_single...)
	}
	return results, nil
}

func (self *default_cti_policy) updateById(table *types.TableDefinition, id interface{},
	updated_attributes map[string]interface{}) error {
	e := self.simple.updateById(table, id, updated_attributes)
	if nil == e {
		return nil
	}
	if e != sql.ErrNoRows {
		return e
	}

	for _, child := range table.OwnChildren.All() {
		e = self.cti.updateById(child, id, updated_attributes)
		if nil == e {
			return nil
		}

		if e != sql.ErrNoRows {
			return e
		}
	}

	return sql.ErrNoRows
}

func (self *default_cti_policy) update(table *types.TableDefinition,
	params map[string]string, updated_attributes map[string]interface{}) (int64, error) {
	effected_single, e := self.simple.update(table, params, updated_attributes)
	if nil != e {
		return 0, e
	}
	effected_all := effected_single

	for _, child := range table.OwnChildren.All() {
		effected_single, e = self.cti.update(child, params, updated_attributes)
		if nil != e {
			return 0, e
		}
		effected_all += effected_single
	}
	return effected_all, nil
}

func (self *default_cti_policy) deleteById(table *types.TableDefinition, id interface{}) error {
	e := self.simple.deleteById(table, id)
	if nil == e {
		return nil
	}

	if e != sql.ErrNoRows {
		return e
	}

	for _, child := range table.OwnChildren.All() {
		e = self.cti.deleteById(child, id)
		if nil == e {
			return nil
		}

		if e != sql.ErrNoRows {
			return e
		}
	}

	return sql.ErrNoRows
}

func (self *default_cti_policy) delete(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	effected_single, e := self.simple.delete(table, params)
	if nil != e {
		return 0, e
	}
	effected_all := effected_single

	for _, child := range table.OwnChildren.All() {
		effected_single, e = self.cti.delete(child, params)
		if nil != e {
			return 0, e
		}
		effected_all += effected_single
	}
	return effected_all, nil
}

func (self *default_cti_policy) forEach(table *types.TableDefinition, params map[string]string,
	cb func(table *types.TableDefinition, id interface{}) error) error {
	e := self.simple.forEach(table, params, cb)
	if nil != e {
		return e
	}

	for _, child := range table.OwnChildren.All() {
		e = self.cti.forEach(child, params, cb)
		if nil != e {
			return e
		}
	}
	return nil
}

// type postgresql_cti_policy struct {
// 	default_cti_policy
// }

// func (self *postgresql_cti_policy) findById(table *types.TableDefinition,
// 	id interface{}, includes string) (map[string]interface{}, error) {

// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT ")
// 	columns := toColumns(table, false)
// 	buffer.WriteString("tableoid::regclass as tablename, ")
// 	if nil == columns || 0 == len(columns) {
// 		return nil, errors.New("crazy! selected columns is empty.")
// 	}
// 	writeColumns(columns, &buffer)
// 	buffer.WriteString(" FROM ")
// 	buffer.WriteString(table.CollectionName)
// 	buffer.WriteString(" WHERE ")
// 	buffer.WriteString(table.Id.Name)
// 	if self.isNumericParams {
// 		buffer.WriteString(" = $1")
// 	} else {
// 		buffer.WriteString(" = ?")
// 	}

// 	new_columns := make([]*types.ColumnDefinition, len(columns)+1)
// 	new_columns[0] = tablename_column
// 	copy(new_columns[1:], columns)

// 	values, e := self.drv.selectOne(buffer.String(), []interface{}{id}, new_columns...)
// 	if nil != e {
// 		return nil, e
// 	}

// 	tablename, ok := values[0].(string)
// 	if !ok {
// 		return nil, errors.New("table name is not a string")
// 	}
// 	if tablename != table.CollectionName {
// 		new_table := table.FindByTableName(tablename)
// 		if nil == new_table {
// 			return nil, errors.New("table name '" + tablename + "' is undefined.")
// 		}
// 		return self.findById(new_table, id, includes)
// 	}
// 	res := make(map[string]interface{})
// 	res["type"] = table.UnderscoreName
// 	for i := 1; i < len(new_columns); i++ {
// 		res[new_columns[i].Name] = values[i]
// 	}

// 	if 0 != len(includes) {
// 		e = self.loadIncludes(table, res, includes)
// 		if nil != e {
// 			return nil, e
// 		}
// 	}
// 	return res, nil
// }

// func (self *postgresql_cti_policy) find(table *types.TableDefinition,
// 	params map[string]string) ([]map[string]interface{}, error) {

// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT tableoid::regclass as tablename, id as id FROM ")
// 	buffer.WriteString(table.CollectionName)
// 	args, e := self.whereWithParams(table, false, 1, params, &buffer)
// 	if nil != e {
// 		return nil, e
// 	}

// 	id_list, e := self.drv.selectAll(buffer.String(), args, class_table_inherit_columns)
// 	if nil != e {
// 		return nil, e
// 	}

// 	var last_builder QueryBuilder
// 	var last_name string

// 	results := make([]map[string]interface{}, 0, len(id_list))
// 	for _, values := range id_list {
// 		tablename, ok := values[0]
// 		if !ok {
// 			panic("'tablename' is not found")
// 		}

// 		name, ok := tablename.(string)
// 		if !ok {
// 			panic("'tablename' is not a string")
// 		}

// 		id, ok := values[1]
// 		if !ok {
// 			panic("'tablename' is not found")
// 		}

// 		if last_name != name {
// 			new_table := table.FindByTableName(name)
// 			if nil == new_table {
// 				return nil, errors.New("table '" + name + "' is undefined.")
// 			}

// 			last_builder, e = self.drv.where(new_table, "id = ?")
// 			if nil != e {
// 				return nil, e
// 			}
// 			last_name = name
// 		}

// 		if nil == last_builder {
// 			return nil, errors.New("table '" + name + "' is undefined.")
// 		}

// 		instance, e := last_builder.Bind(id).Build().One()
// 		if nil != e {
// 			return nil, e
// 		}

// 		results = append(results, instance)
// 	}
// 	return results, nil
// }
