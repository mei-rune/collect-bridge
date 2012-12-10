package mdb

type ObjectId struct {
	definition *ClassDefinition
	id         string
}
type Driver interface {
	Insert(cd *ClassDefinition, properties map[string]interface{}, parents []ObjectId) (interface{}, error)
	Update(cd *ClassDefinition, properties map[string]interface{}, parents []ObjectId) error
	FindById(cd *ClassDefinition, id string, parents []ObjectId) (interface{}, error)
	Delete(cd *ClassDefinition, id string, parents []ObjectId) error
}
