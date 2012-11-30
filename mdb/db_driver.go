package mdb

type Driver interface {
	Insert(tableName string, properties map[string]interface{}) (uint64, error)
	Update(tableName string, properties map[string]interface{}, condition string) (uint64, error)
	Get(condition interface{}, args ...interface{}) (interface{}, error)
  Delete(tableName string, output interface{}) (int64, error) {
}
