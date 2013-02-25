package mdb

import (
	"bytes"
)

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
	TargetClass   *ClassDefinition
	ForeignKey    string
	AttributeName string
	Embedded      bool
	Polymorphic   bool
}

func (self *HasMany) Type() AssocationType {
	return HAS_MANG
}

func (self *HasMany) Target() *ClassDefinition {
	return self.TargetClass
}

type HasOne struct {
	TargetClass   *ClassDefinition
	ForeignKey    string
	AttributeName string
	Embedded      bool
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
	IsReadOnly   bool
	IsUniquely   bool
	Restrictions []Validator
	DefaultValue interface{}
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

func (self *ClassDefinition) RootClass() *ClassDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}

func (self *ClassDefinition) IsInheritance() bool {
	return (nil != self.Super) || (nil != self.Children && 0 != len(self.Children))
}
func (self *ClassDefinition) InheritanceFrom(cls *ClassDefinition) bool {
	for s := self; nil != s; s = s.Super {
		if s == cls {
			return true
		}
	}
	return false
}
func (self *ClassDefinition) CollectionName() string {
	if nil == self.Super {
		return self.collectionName
	}

	return self.Super.CollectionName()
}

func (self *ClassDefinition) GetAssocationByCollectionName(nm string) Assocation {
	if nil == self.Assocations {
		return nil
	}
	for _, assoc := range self.Assocations {
		if nm == assoc.Target().CollectionName() {
			return assoc
		}
	}
	return nil
}

func (self *ClassDefinition) GetProperty(nm string) (pr *PropertyDefinition, ok bool) {
	pr, ok = self.Properties[nm]
	return pr, ok
}

func (self *ClassDefinition) GetOwnProperty(nm string) (pr *PropertyDefinition, ok bool) {
	pr, ok = self.OwnProperties[nm]
	return pr, ok
}

func (self *ClassDefinition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("class ")
	buffer.WriteString(self.Name)
	if nil != self.Super {
		buffer.WriteString(" < ")
		buffer.WriteString(self.Super.Name)
		buffer.WriteString(" { ")
	} else {
		buffer.WriteString(" { ")
	}
	if nil != self.OwnProperties && 0 != len(self.OwnProperties) {
		for _, pr := range self.OwnProperties {
			buffer.WriteString(pr.Name)
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString(" }")
	return buffer.String()
}
