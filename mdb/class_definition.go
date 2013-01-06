package mdb

import (
//"commons/stringutils"
)

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
	ForeignKey  string
}

func (self *HasMany) Type() AssocationType {
	return BELONGS_TO
}

func (self *HasMany) Target() *ClassDefinition {
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
	Super          *ClassDefinition
	Name           string
	collectionName string

	OwnProperties map[string]*PropertyDefinition
	Properties    map[string]*PropertyDefinition
	Assocations   []Assocation
	Children      []*ClassDefinition
}

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
