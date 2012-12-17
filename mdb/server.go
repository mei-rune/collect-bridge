package mdb

import (
	"errors"
)

type MdbServer struct {
	restrict    bool
	driver      Driver
	definitions *ClassDefinitions
}

func (self *MdbServer) convert(cls *ClassDefinition, attributes map[string]interface{}, is_update bool) (map[string]interface{}, error) {
	new_attributes := make(map[string]interface{}, len(attributes))
	errs := make([]error, 0, 10)
	for k, pr := range cls.Properties {
		var new_value interface{}
		value, ok := attributes[k]
		if !ok {
			if is_update {
				continue
			}

			if pr.IsRequired {
				errs = append(errs, errors.New("'"+k+"' is required"))
				continue
			}
			new_value = pr.DefaultValue
		} else {
			if self.restrict {
				delete(attributes, k)
			}
			var err error
			new_value, err = pr.Type.Convert(value)
			if nil != err {
				errs = append(errs, errors.New("'"+k+"' convert to internal value failed, "+err.Error()))
				continue
			}
		}
		if nil != pr.Restrictions && 0 != len(pr.Restrictions) {
			is_failed := false
			for _, r := range pr.Restrictions {
				if ok, err := r.Validate(new_value, attributes); !ok {
					errs = append(errs, errors.New("'"+k+"' is validate failed, "+err.Error()))
					is_failed = true
				}
			}

			if is_failed {
				continue
			}
		}

		new_attributes[k] = new_value
	}

	if 0 != len(errs) {
		return nil, &MutiErrors{msg: "validate failed", errs: errs}
	}
	if self.restrict && 0 != len(attributes) {
		for k, _ := range attributes {
			errs = append(errs, errors.New("'"+k+"' is useless"))
		}
		return nil, &MutiErrors{msg: "validate failed", errs: errs}
	}
	return new_attributes, nil
}

func (self *MdbServer) Create(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	new_attributes, errs := self.convert(cls, attributes, false)
	if nil != errs {
		return nil, errs
	}

	return self.driver.Insert(cls, new_attributes)
}
func (self *MdbServer) FindById(cls *ClassDefinition, id string) (interface{}, error) {
	return self.driver.FindById(cls, id)
}

func (self *MdbServer) Update(cls *ClassDefinition, id string, attributes map[string]interface{}) error {
	new_attributes, errs := self.convert(cls, attributes, true)
	if nil != errs {
		return errs
	}

	return self.driver.Update(cls, id, new_attributes)
}

func (self *MdbServer) RemoveById(cls *ClassDefinition, id string) error {
	return self.driver.Delete(cls, id)
}
