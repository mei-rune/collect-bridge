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
	OwnChildren    *TableDefinitions
	Children       *TableDefinitions
	Assocations    []Assocation
}

func (self *TableDefinition) Root() *TableDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}

func (self *TableDefinition) IsSubclassOf(super *TableDefinition) bool {
	for s := self; nil != s; s = s.Super {
		if s == super {
			return true
		}
	}
	return false
}

func (self *TableDefinition) IsSingleTableInheritance() bool {
	_, ok := self.Attributes["type"]
	return ok
}

func (self *TableDefinition) HasChildren() bool {
	return (nil != self.OwnChildren && !self.OwnChildren.IsEmpty())
}

func (self *TableDefinition) IsInheritanced() bool {
	return (nil != self.Super) || (nil != self.OwnChildren && !self.OwnChildren.IsEmpty())
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

func (self *TableDefinition) FindByUnderscoreName(nm string) *TableDefinition {
	if self.UnderscoreName == nm {
		return self
	}
	if !self.HasChildren() {
		return nil
	}
	return self.Children.FindByUnderscoreName(nm)
}

func (self *TableDefinition) FindByTableName(nm string) *TableDefinition {
	if self.CollectionName == nm {
		return self
	}
	if !self.HasChildren() {
		return nil
	}
	return self.Children.FindByTableName(nm)
}

func (self *TableDefinition) Find(nm string) *TableDefinition {
	if self.UnderscoreName == nm {
		return self
	}
	if !self.HasChildren() {
		return nil
	}
	return self.Children.Find(nm)
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
	table2definitions      map[string]*TableDefinition
}

func NewTableDefinitions() *TableDefinitions {
	return &TableDefinitions{underscore2Definitions: make(map[string]*TableDefinition),
		definitions:       make(map[string]*TableDefinition),
		table2definitions: make(map[string]*TableDefinition)}
}

func (self *TableDefinitions) FindByUnderscoreName(nm string) *TableDefinition {
	return self.underscore2Definitions[nm]
}

func (self *TableDefinitions) FindByTableName(nm string) *TableDefinition {
	return self.table2definitions[nm]
}

func (self *TableDefinitions) Find(nm string) *TableDefinition {
	return self.definitions[nm]
}

func stiRoot(cls *TableDefinition) *TableDefinition {
	for s := cls; ; s = s.Super {
		if nil == s.Super {
			return s
		}
		if s.Super.CollectionName != cls.CollectionName {
			return s
		}
	}
}

func (self *TableDefinitions) Register(cls *TableDefinition) {
	self.definitions[cls.Name] = cls
	self.underscore2Definitions[cls.UnderscoreName] = cls
	if table, ok := self.table2definitions[cls.CollectionName]; ok {
		if table.IsSubclassOf(cls) {
			// self.table2definitions[cls.CollectionName] = cls
		} else if stiRoot(cls) != stiRoot(table) {
			panic("table '" + cls.Name + "' and table '" + table.Name + "' is same with collection name.")
		}
	} else {
		self.table2definitions[cls.CollectionName] = cls
	}
}

func (self *TableDefinitions) Unregister(cls *TableDefinition) {
	delete(self.definitions, cls.Name)
	delete(self.underscore2Definitions, cls.UnderscoreName)
	// tables := self.table2definitions[cls.CollectionName]
	// if nil != tables {
	// 	delete(tables, cls.UnderscoreName)
	// 	if 0 == len(tables) {
	// 		delete(self.table2definitions, cls.CollectionName)
	// 	}
	// }
}

func (self *TableDefinitions) All() map[string]*TableDefinition {
	return self.definitions
}

func (self *TableDefinitions) Len() int {
	return len(self.definitions)
}

func (self *TableDefinitions) IsEmpty() bool {
	return 0 == len(self.definitions)
}

func (self *TableDefinitions) UnderscoreAll() map[string]*TableDefinition {
	return self.underscore2Definitions
}
