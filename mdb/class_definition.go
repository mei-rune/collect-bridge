package mdb

type AssocationType int

const (
	BELONGS_TO              AssocationType = 1
	HAS_MANG                AssocationType = 2
	HAS_AND_BELONGS_TO_MANY AssocationType = 3
)

type Assocation interface {
	Type() AssocationType
	Target() *ClassDefinition
}

type BelongsTo struct {
	TargetClass *ClassDefinition
	Name        *PropertyDefinition
}

func (self *BelongsTo) Type() AssocationType {
	return BELONGS_TO
}

func (self *BelongsTo) Target() *ClassDefinition {
	return self.TargetClass
}

type HasMany struct {
	TargetClass *ClassDefinition
}

func (self *HasMany) Type() AssocationType {
	return BELONGS_TO
}

func (self *HasMany) Target() *ClassDefinition {
	return self.TargetClass
}

type HasAndBelongsToMany struct {
	TargetClass *ClassDefinition
}

func (self *HasAndBelongsToMany) Type() AssocationType {
	return BELONGS_TO
}

func (self *HasAndBelongsToMany) Target() *ClassDefinition {
	return self.TargetClass
}

type PropertyDefinition struct {
	Name         string
	Type         TypeDefinition
	restriction  []Validator
	defaultValue interface{}
}

type MutiErrors struct {
	msg  string
	errs []error
}

func (self *MutiErrors) Error() string {
	return self.msg
}
func (self *MutiErrors) Errors() []error {
	return self.errs
}

func (self *PropertyDefinition) Validate(obj interface{}) (bool, error) {
	if nil == self.restriction {
		return true, nil
	}

	var result bool = true
	var errs []error = make([]error, 0, len(self.restriction))
	for _, Validator := range self.restriction {
		if ok, err := Validator.Validate(obj); !ok {
			result = false
			errs = append(errs, err)
		}
	}

	if result {
		return true, nil
	}
	return false, &MutiErrors{errs: errs, msg: "property '" + self.Name + "' is error"}
}

type ClassDefinition struct {
	Super *ClassDefinition
	Name  string

	ownProperties map[string]*PropertyDefinition

	properties  map[string]*PropertyDefinition
	assocations []Assocation
}

func (self *ClassDefinition) CollectionName() string {
	if nil == self.Super {
		return self.Name
	}

	return self.Super.CollectionName()
}

func (self *ClassDefinition) GetProperty(nm string) (pr *PropertyDefinition, ok bool) {
	pr, ok = self.properties[nm]
	return pr, ok
}

func (self *ClassDefinition) GetOwnProperty(nm string) (pr *PropertyDefinition, ok bool) {
	pr, ok = self.ownProperties[nm]
	return pr, ok
}
