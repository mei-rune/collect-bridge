package types

import (
	"bytes"
)

type CollectionType int

const (
	COLLECTION_UNKNOWN CollectionType = 0
	COLLECTION_ARRAY   CollectionType = 1
	COLLECTION_SET     CollectionType = 2
)

func (t CollectionType) IsArray() bool {
	return t == COLLECTION_ARRAY
}

func (t CollectionType) IsSet() bool {
	return t == COLLECTION_SET
}

func (t CollectionType) IsCollection() bool {
	return t == COLLECTION_SET || t == COLLECTION_ARRAY
}

type AttributeDefinition struct {
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
	Super          *ClassDefinition
	Name           string
	UnderscoreName string
	OwnAttributes  map[string]*AttributeDefinition
	Attributes     map[string]*AttributeDefinition
	Children       []*ClassDefinition
}

func (self *ClassDefinition) Root() *ClassDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}

func (self *ClassDefinition) IsSubclassOf(cls *ClassDefinition) bool {
	s := self
	for nil != s {
		if s == cls {
			return true
		}

		s = s.Super
	}
	return false
}

func (self *ClassDefinition) IsInheritanced() bool {
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

func (self *ClassDefinition) GetAttribute(nm string) (pr *AttributeDefinition) {
	return self.Attributes[nm]
}

func (self *ClassDefinition) GetAttributes() map[string]*AttributeDefinition {
	return self.Attributes
}

func (self *ClassDefinition) GetOwnAttribute(nm string) (pr *AttributeDefinition) {
	return self.OwnAttributes[nm]
}

func (self *ClassDefinition) GetOwnAttributes() map[string]*AttributeDefinition {
	return self.OwnAttributes
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

type ClassDefinitions struct {
	underscore2Definitions map[string]*ClassDefinition
	clsDefinitions         map[string]*ClassDefinition
}

func (self *ClassDefinitions) FindByUnderscoreName(nm string) *ClassDefinition {
	return self.underscore2Definitions[nm]
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	return self.clsDefinitions[nm]
}

func (self *ClassDefinitions) Register(cls *ClassDefinition) {
	self.clsDefinitions[cls.Name] = cls
	self.underscore2Definitions[cls.UnderscoreName] = cls
}

func (self *ClassDefinitions) Unregister(cls *ClassDefinition) {
	delete(self.clsDefinitions, cls.Name)
	delete(self.underscore2Definitions, cls.UnderscoreName)
}

func (self *ClassDefinitions) All(cls *ClassDefinition) map[string]*ClassDefinition {
	return self.clsDefinitions
}
