package ds

// import (
// 	"commons/types"
// )

// type default_cti_policy struct {
// 	drv driver
// }

// func (self *default_cti_policy) count(table *types.TableDefinition,
// 	params map[string]string) (int64, error) {
// 	effected_single, e := self.drv.countInSpecificTable(table, params)
// 	if nil != e {
// 		return 0, e
// 	}
// 	effected_all := effected_single

// 	for _, child := range table.OwnChildren.All() {
// 		effected_single, e := self.drv.count(child, params)
// 		if nil != e {
// 			return 0, e
// 		}
// 		effected_all += effected_single
// 	}
// 	return effected_all, nil
// }

// func (self *default_cti_policy) findById(table *types.TableDefinition,
// 	id interface{}, includes string) (map[string]interface{}, error) {
// 	result, e := self.drv.findByIdInSpecificTable(table, id, includes)
// 	if nil == e {
// 		return result, nil
// 	}
// 	if e != sql.ErrNoRows {
// 		return nil, e
// 	}

// 	for _, child := range table.OwnChildren.All() {
// 		result, e = self.drv.findById(child, id, includes)
// 		if nil == e {
// 			return result, nil
// 		}

// 		if e != sql.ErrNoRows {
// 			return nil, e
// 		}
// 	}

// 	return nil, sql.ErrNoRows
// }

// func (self *default_cti_policy) find(table *types.TableDefinition,
// 	params map[string]string) ([]map[string]interface{}, error) {
// 	results, e := self.drv.findInSpecificTable(table, params)
// 	if nil != e && e != sql.ErrNoRows {
// 		return nil, e
// 	}

// 	for _, child := range table.OwnChildren.All() {
// 		results_single, e := self.drv.find(child, params)
// 		if nil != e && e != sql.ErrNoRows {
// 			return nil, e
// 		}
// 		results = append(results, results_single...)
// 	}
// 	return results, nil
// }

// func (self *default_cti_policy) updateById(table *types.TableDefinition, id interface{},
// 	updated_attributes map[string]interface{}) error {
// 	e := self.drv.updateByIdInSpecificTable(table, id, updated_attributes)
// 	if nil == e {
// 		return nil
// 	}
// 	if e != sql.ErrNoRows {
// 		return e
// 	}

// 	for _, child := range table.OwnChildren.All() {
// 		e = self.drv.updateById(child, id, updated_attributes)
// 		if nil == e {
// 			return nil
// 		}

// 		if e != sql.ErrNoRows {
// 			return e
// 		}
// 	}

// 	return sql.ErrNoRows
// }

// func (self *default_cti_policy) update(table *types.TableDefinition,
// 	params map[string]string,
// 	updated_attributes map[string]interface{}) (int64, error) {
// 	effected_single, e := self.drv.updateInSpecificTable(table, params, updated_attributes)
// 	if nil != e {
// 		return 0, e
// 	}
// 	effected_all := effected_single

// 	for _, child := range table.OwnChildren.All() {
// 		effected_single, e = self.drv.update(child, params, updated_attributes)
// 		if nil != e {
// 			return 0, e
// 		}
// 		effected_all += effected_single
// 	}
// 	return effected_all, nil
// }

// func (self *default_cti_policy) deleteById(table *types.TableDefinition, id interface{}) error {
// 	e := self.drv.deleteByIdInSpecificTable(table, id)
// 	if nil == e {
// 		return nil
// 	}

// 	if e != sql.ErrNoRows {
// 		return e
// 	}

// 	for _, child := range table.OwnChildren.All() {
// 		e = self.drv.deleteById(child, id)
// 		if nil == e {
// 			return nil
// 		}

// 		if e != sql.ErrNoRows {
// 			return e
// 		}
// 	}

// 	return sql.ErrNoRows
// }

// func (self *default_cti_policy) delete(table *types.TableDefinition,
// 	params map[string]string) (int64, error) {
// 	effected_single, e := self.drv.deleteInSpecificTable(table, params)
// 	if nil != e {
// 		return 0, e
// 	}
// 	effected_all := effected_single

// 	for _, child := range table.OwnChildren.All() {
// 		effected_single, e = self.drv.delete(child, params)
// 		if nil != e {
// 			return 0, e
// 		}
// 		effected_all += effected_single
// 	}
// 	return effected_all, nil
// }
