package mdb

import ()

//"commons/stringutils"
type CollectionType int
type AssocationType int

const (
	BELONGS_TO              AssocationType = 1
	HAS_MANG                AssocationType = 2
	HAS_ONE                 AssocationType = 3
	HAS_AND_BELONGS_TO_MANY AssocationType = 4

	COLLECTION_UNKNOWN CollectionType = 0
	COLLECTION_ARRAY   CollectionType = 1
	COLLECTION_SET     CollectionType = 2
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
	ForeignKey  string
}

func (self *HasMany) Type() AssocationType {
	return HAS_MANG
}

func (self *HasMany) Target() *ClassDefinition {
	return self.TargetClass
}

type HasOne struct {
	TargetClass   *ClassDefinition
	AttributeName string
}

func (self *HasOne) Type() AssocationType {
	return HAS_ONE
}

func (self *HasOne) Target() *ClassDefinition {
	return self.TargetClass
}

type HasAndBelongsToMany struct {
	TargetClass    *ClassDefinition
	Through        *ClassDefinition
	CollectionName string
	ForeignKey     string
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
	Collection   CollectionType
	IsRequired   bool
	Restrictions []Validator
	DefaultValue interface{}
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

type ClassDefinition struct {
	Super            *ClassDefinition
	Name             string
	collectionName   string
	HierarchicalType *HierarchicalEnumeration
	OwnProperties    map[string]*PropertyDefinition
	Properties       map[string]*PropertyDefinition
	Assocations      []Assocation
	Children         []*ClassDefinition
}

type HierarchicalEnumeration struct {
	Value, MinValue, MaxValue int
}

// func (self *ClassDefinition) String() string {
// 	return fmt.Sprintf(`class %s : %s {
// 	CollectionName  %s
// 	table  %s` 
// 	self.Name, self.Super.Name, self.collectionName, self.CollectionName(),
// 	self.HierarchicalType.Value, self.HierarchicalType.MinValue, self.HierarchicalType.MaxValue,
// 	OwnProperties    map[string]*PropertyDefinition
// 	Properties       map[string]*PropertyDefinition
// 	Assocations      []Assocation
// 	Children         []*ClassDefinition)
// }

func (self *ClassDefinition) CollectionName() string {
	if nil == self.Super {
		return self.collectionName
	}

	return self.Super.CollectionName()
}

func (self *ClassDefinition) GetProperty(nm string) (pr *PropertyDefinition, ok bool) {
	pr, ok = self.Properties[nm]
	return pr, ok
}

func (self *ClassDefinition) GetOwnProperty(nm string) (pr *PropertyDefinition, ok bool) {
	pr, ok = self.OwnProperties[nm]
	return pr, ok
}
