package data_store

import (
	"commons/types"
)

type cti_supported_driver struct {
	simple     driver
	cti_policy driver
}

func ctiSupportWithPostgreSQLInherit(simple *simple_driver) *cti_supported_driver {
	cti_policy := &postgresql_cti_policy{simple: simple}
	res := &cti_supported_driver{simple: simple, cti_policy: cti_policy}
	//cti_policy.cti = res
	return res
}

func ctiSupport(simple driver) *cti_supported_driver {
	cti_policy := &default_cti_policy{simple: simple}
	res := &cti_supported_driver{simple: simple, cti_policy: cti_policy}
	cti_policy.cti = res
	return res
}

func (self *cti_supported_driver) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	return self.simple.insert(table, attributes)
}

func (self *cti_supported_driver) count(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	if table.IsSingleTableInheritance() {
		return self.simple.count(table, params)
	}

	if !table.HasChildren() {
		return self.simple.count(table, params)
	}

	return self.cti_policy.count(table, params)
}

func (self *cti_supported_driver) snapshot(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {
	if table.IsSingleTableInheritance() {
		return self.simple.snapshot(table, params)
	}

	if !table.HasChildren() {
		return self.simple.snapshot(table, params)
	}

	return self.cti_policy.snapshot(table, params)
}

func (self *cti_supported_driver) findById(table *types.TableDefinition, id interface{}) (map[string]interface{}, error) {
	if table.IsSingleTableInheritance() {
		return self.simple.findById(table, id)
	}

	if !table.HasChildren() {
		return self.simple.findById(table, id)
	}

	return self.cti_policy.findById(table, id)
}

func (self *cti_supported_driver) find(table *types.TableDefinition, params map[string]string) ([]map[string]interface{}, error) {
	if table.IsSingleTableInheritance() {
		return self.simple.find(table, params)
	}

	if !table.HasChildren() {
		return self.simple.find(table, params)
	}

	return self.cti_policy.find(table, params)
}

func (self *cti_supported_driver) updateById(table *types.TableDefinition, id interface{},
	updated_attributes map[string]interface{}) error {
	if table.IsSingleTableInheritance() {
		return self.simple.updateById(table, id, updated_attributes)
	}

	if !table.HasChildren() {
		return self.simple.updateById(table, id, updated_attributes)
	}

	return self.cti_policy.updateById(table, id, updated_attributes)
}

func (self *cti_supported_driver) update(table *types.TableDefinition, params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {
	if table.IsSingleTableInheritance() {
		return self.simple.update(table, params, updated_attributes)
	}

	if !table.HasChildren() {
		return self.simple.update(table, params, updated_attributes)
	}

	return self.cti_policy.update(table, params, updated_attributes)
}

func (self *cti_supported_driver) deleteById(table *types.TableDefinition, id interface{}) error {
	if table.IsSingleTableInheritance() {
		return self.simple.deleteById(table, id)
	}

	if !table.HasChildren() {
		return self.simple.deleteById(table, id)
	}

	return self.cti_policy.deleteById(table, id)
}

func (self *cti_supported_driver) delete(table *types.TableDefinition, params map[string]string) (int64, error) {
	if table.IsSingleTableInheritance() {
		return self.simple.delete(table, params)
	}

	if !table.HasChildren() {
		return self.simple.delete(table, params)
	}

	return self.cti_policy.delete(table, params)
}

func (self *cti_supported_driver) forEach(table *types.TableDefinition, params map[string]string,
	cb func(table *types.TableDefinition, id interface{}) error) error {

	if table.IsSingleTableInheritance() {
		return self.simple.forEach(table, params, cb)
	}

	if !table.HasChildren() {
		return self.simple.forEach(table, params, cb)
	}

	return self.cti_policy.forEach(table, params, cb)
}
