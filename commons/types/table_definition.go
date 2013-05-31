package types

import (
	"bytes"
)

type AssocationType int

const (
	BELONGS_TO              AssocationType = 1
	HAS_MANG                AssocationType = 2
	HAS_ONE                 AssocationType = 3
	HAS_AND_BELONGS_TO_MANY AssocationType = 4
)

type Assocation interface {
	Type() AssocationType
	Target() *TableDefinition
}

type BelongsTo struct {
	TargetTable *TableDefinition
	Name        *ColumnDefinition
}

func (self *BelongsTo) Type() AssocationType {
	return BELONGS_TO
}

func (self *BelongsTo) Target() *TableDefinition {
	return self.TargetTable
}

type HasMany struct {
	TargetTable *TableDefinition
	ForeignKey  string
	Polymorphic bool
}

func (self *HasMany) Type() AssocationType {
	return HAS_MANG
}

func (self *HasMany) Target() *TableDefinition {
	return self.TargetTable
}

type HasOne struct {
	TargetTable *TableDefinition
	ForeignKey  string
	Polymorphic bool
}

func (self *HasOne) Type() AssocationType {
	return HAS_ONE
}

func (self *HasOne) Target() *TableDefinition {
	return self.TargetTable
}

type HasAndBelongsToMany struct {
	TargetTable *TableDefinition
	Through     *TableDefinition
	ForeignKey  string
}

func (self *HasAndBelongsToMany) Type() AssocationType {
	return BELONGS_TO
}

func (self *HasAndBelongsToMany) Target() *TableDefinition {
	return self.TargetTable
}

type ColumnDefinition struct {
	AttributeDefinition
}

func (self *ColumnDefinition) IsSerial() bool {
	return "id" == self.Name
}

func (self *ColumnDefinition) IsPromaryKey() bool {
	return "id" == self.Name
}

type TableDefinition struct {
	Super          *TableDefinition
	Name           string
	UnderscoreName string
	CollectionName string
	Id             *ColumnDefinition
	OwnAttributes  map[string]*ColumnDefinition
	Attributes     map[string]*ColumnDefinition
	Children       []*TableDefinition
	Assocations    []Assocation
}

func (self *TableDefinition) Root() *TableDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}

func (self *TableDefinition) IsSubclassOf(cls *TableDefinition) bool {
	s := self
	for nil != s {
		if s == cls {
			return true
		}

		s = s.Super
	}
	return false
}

func (self *TableDefinition) IsInheritanced() bool {
	return (nil != self.Super) || (nil != self.Children && 0 != len(self.Children))
}

func (self *TableDefinition) InheritanceFrom(cls *TableDefinition) bool {
	for s := self; nil != s; s = s.Super {
		if s == cls {
			return true
		}
	}
	return false
}

func (self *TableDefinition) GetAttribute(nm string) (pr *ColumnDefinition) {
	return self.Attributes[nm]
}

func (self *TableDefinition) GetAttributes() map[string]*ColumnDefinition {
	return self.Attributes
}

func (self *TableDefinition) GetOwnAttribute(nm string) (pr *ColumnDefinition) {
	return self.OwnAttributes[nm]
}

func (self *TableDefinition) GetOwnAttributes() map[string]*ColumnDefinition {
	return self.OwnAttributes
}

func (self *TableDefinition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("table ")
	buffer.WriteString(self.Name)
	if nil != self.Super {
		buffer.WriteString(" < ")
		buffer.WriteString(self.Super.Name)
		buffer.WriteString(" { ")
	} else {
		buffer.WriteString(" { ")
	}
	if nil != self.OwnAttributes && 0 != len(self.OwnAttributes) {
		for _, pr := range self.OwnAttributes {
			buffer.WriteString(pr.Name)
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString(" }")
	return buffer.String()
}

func (self *TableDefinition) GetAssocationByTarget(cls *TableDefinition) Assocation {
	if nil != self.Assocations {
		for _, assoc := range self.Assocations {
			if cls.IsSubclassOf(assoc.Target()) {
				return assoc
			}
		}
	}
	if nil != self.Super {
		return self.Super.GetAssocationByTarget(cls)
	}
	return nil
}

func (self *TableDefinition) GetAssocationByTargetAndTypes(cls *TableDefinition,
	assocationTypes ...AssocationType) Assocation {

	if nil != self.Assocations {
		for _, assoc := range self.Assocations {
			found := false
			for _, assocationType := range assocationTypes {
				if assocationType == assoc.Type() {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if cls.IsSubclassOf(assoc.Target()) {
				return assoc
			}
		}
	}
	if nil != self.Super {
		return self.Super.GetAssocationByTargetAndTypes(cls, assocationTypes...)
	}
	return nil
}

type TableDefinitions struct {
	underscore2Definitions map[string]*TableDefinition
	definitions            map[string]*TableDefinition
}

func (self *TableDefinitions) FindByUnderscoreName(nm string) *TableDefinition {
	return self.underscore2Definitions[nm]
}

func (self *TableDefinitions) Find(nm string) *TableDefinition {
	return self.definitions[nm]
}

func (self *TableDefinitions) Register(cls *TableDefinition) {
	self.definitions[cls.Name] = cls
	self.underscore2Definitions[cls.UnderscoreName] = cls
}

func (self *TableDefinitions) Unregister(cls *TableDefinition) {
	delete(self.definitions, cls.Name)
	delete(self.underscore2Definitions, cls.UnderscoreName)
}

func (self *TableDefinitions) All(cls *TableDefinition) map[string]*TableDefinition {
	return self.definitions
}
