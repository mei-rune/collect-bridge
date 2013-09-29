package data_store

import (
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-sqlite3"
	"strconv"
	"strings"
)

type session struct {
	simple *simple_driver
	drv    driver
}

func newSession(drvName string, conn *sql.DB, tables *types.TableDefinitions) *session {
	simple := simpleDriver(drvName, conn, !*isPostgresqlInherit, tables)
	if POSTGRESQL == simple.dbType && *isPostgresqlInherit {
		return &session{simple: simple, drv: ctiSupportWithPostgreSQLInherit(simple)}
	}
	return &session{simple: simple, drv: ctiSupport(simple)}
}

func (self *session) count(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	return self.drv.count(table, params)
}

func (self *session) snapshot(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {
	return self.drv.snapshot(table, params)
}

func (self *session) findById(table *types.TableDefinition, id interface{}, includes string) (map[string]interface{}, error) {
	result, e := self.drv.findById(table, id)
	if nil != e {
		return nil, e
	}

	if 0 != len(includes) {
		e = self.loadIncludes(table, result, includes)
		if nil != e && sql.ErrNoRows != e {
			return nil, e
		}
	}
	return result, nil
}

func (self *session) find(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {

	results, e := self.drv.find(table, params)
	if nil != e {
		return nil, e
	}

	if includes, ok := params["includes"]; ok && 0 != len(includes) {
		for _, result := range results {
			e = self.loadIncludes(table, result, includes)
			if nil != e && sql.ErrNoRows != e {
				return nil, e
			}
		}
	}
	return results, e
}

func (self *session) loadIncludes(parent_table *types.TableDefinition, parent map[string]interface{}, includes string) error {
	parent_id := parent["id"]
	if nil == parent_id {
		return errors.New("parent id is nil while load children.")
	}

	if "*" == includes {
		assocations := parent_table.GetAssocationByTypes(types.HAS_ONE, types.HAS_MANY)

		if nil == assocations || 0 == len(assocations) {
			return nil
		}
		for _, assocation := range assocations {
			results, e := self.findByParent(parent_table, parent_id, assocation, assocation.Target())
			if nil != e {
				return e
			}
			parent["$"+assocation.Target().UnderscoreName] = results
		}
	} else {
		for _, s := range strings.Split(includes, ",") {
			target := self.simple.tables.FindByUnderscoreName(s)
			if nil == target {
				return errors.New("table '" + s + "' is not found in the includes.")
			}
			assocations := parent_table.GetAssocationByTargetAndTypes(target, types.HAS_ONE, types.HAS_MANY)
			if nil == assocations || 0 == len(assocations) {
				return errors.New("assocation that to '" + s + "' is not found in the includes.")
			}

			for _, assocation := range assocations {
				results, e := self.findByParent(parent_table, parent_id, assocation, target)
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
	parent_id interface{}, assocation types.Assocation,
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
		params["@parent_id"] = fmt.Sprint(parent_id)
	} else {
		params["@"+foreignKey] = fmt.Sprint(parent_id)
	}
	return self.find(target, params)
}

func (self *session) children(parent_table *types.TableDefinition, parent_id interface{},
	target *types.TableDefinition, foreignKey string) ([]map[string]interface{}, error) {
	assocation, e := parent_table.GetAssocation(target, foreignKey, types.HAS_MANY, types.HAS_ONE)
	if nil != e {
		return nil, e
	}
	return self.findByParent(parent_table, parent_id, assocation, target)
}

func (self *session) parentBy(child_table *types.TableDefinition, child_id interface{},
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

		parent := self.simple.tables.FindByUnderscoreName(parent_type)
		if nil == parent {
			return nil, errors.New(" table '" + parent_type + "' is not exists.")
		}

		if parent != target && parent.IsSubclassOf(target) {
			return nil, errors.New(" table '" + parent_type +
				"' is not a subclass of table '" + target.UnderscoreName + "'.")
		}

		parent_id := res["parent_id"]
		if nil == v {
			return nil, errors.New("'parent_id' is nil in the result")
		}

		return self.findById(parent, parent_id, "")
	}

	id := res[foreignKey]
	if nil == id {
		return nil, errors.New("'" + foreignKey + "' is not exists in the result.")
	}

	return self.findById(target, id, "")

}

func (self *session) parent(child_table *types.TableDefinition, child_id interface{},
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

	return self.findById(target, id, "")
}

////////////////////////// insert //////////////////////////
func (self *session) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	id, e := self.drv.insert(table, attributes)
	if nil != e {
		return 0, e
	}
	e = self.createChildren(table, id, attributes)
	if nil != e {
		return 0, e
	}
	return id, nil
}

func (self *session) save(table *types.TableDefinition, params map[string]string,
	attributes map[string]interface{}) (int, int64, error) {

	results, e := self.find(table, params)
	if nil != e && e != sql.ErrNoRows {
		return -1, 0, e
	}

	if nil == results || 0 == len(results) {
		id, e := self.drv.insert(table, attributes)
		if nil != e {
			return 1, 0, e
		}
		e = self.createChildren(table, id, attributes)
		if nil != e {
			return 1, 0, e
		}
		return 1, id, nil
	} else if 1 == len(results) {
		id := results[0][table.Id.Name]
		e = self.updateById(table, id, attributes)
		if nil != e {
			return 0, 0, e
		}
		switch v := id.(type) {
		case int64:
			return 0, v, nil
		case int:
			return 0, int64(v), nil
		case int32:
			return 0, int64(v), nil
		case string:
			i64, e := strconv.ParseInt(v, 10, 64)
			if nil != e {
				return 0, 0, fmt.Errorf("id is not a int64 - %v", id)
			}
			return 0, i64, nil
		default:
			return 0, 0, fmt.Errorf("id is not a int64 - %v", id)
		}
	} else {
		return 0, 0, fmt.Errorf("results that match condition is not equals 1, actual is %v", len(results))
	}
}

func (self *session) insertByParent(parent_table *types.TableDefinition, parent_id interface{},
	target *types.TableDefinition, foreignKey string, attributes map[string]interface{}) (int64, error) {
	assocation, e := parent_table.GetAssocation(target, foreignKey, types.HAS_MANY, types.HAS_ONE)
	if nil != e {
		return 0, e
	}

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
		return 0, errors.New("unsupported assocation type - " + assocation.Type().String())
	}

	return self.createChild(target, parent_table, parent_id, attributes, foreignKey, is_polymorphic)
}

func (self *session) createChildren(parent_table *types.TableDefinition,
	parent_id interface{}, attributes map[string]interface{}) error {
	for name, v := range attributes {
		if '$' != name[0] {
			continue
		}

		target_table := self.simple.tables.FindByUnderscoreName(name[1:])
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
			} else if values, ok := v.(map[string]interface{}); ok {
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

func (self *session) updateById(table *types.TableDefinition, id interface{},
	updated_attributes map[string]interface{}) error {
	return self.drv.updateById(table, id, updated_attributes)
}

func (self *session) update(table *types.TableDefinition, params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {
	return self.drv.update(table, params, updated_attributes)
}

func (self *session) delete(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	_, e := self.deleteCascadeByParams(table, params)
	if nil != e {
		return 0, e
	}

	return self.drv.delete(table, params)
}

func (self *session) deleteById(table *types.TableDefinition, id interface{}) error {
	_, e := self.deleteCascadeById(table, id)
	if nil != e {
		return e
	}

	return self.drv.deleteById(table, id)
}
