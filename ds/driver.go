package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	//"errors"
	"strconv"
)

//type session interface {
// count(table *types.TableDefinition, params map[string]string) (int64, error)

// findById(table *types.TableDefinition, id interface{}, includes string) (map[string]interface{}, error)
// find(table *types.TableDefinition, params map[string]string) ([]map[string]interface{}, error)

// updateById(table *types.TableDefinition, id interface{}, updated_attributes map[string]interface{}) (int64, error)
// update(table *types.TableDefinition, params map[string]string, updated_attributes map[string]interface{}) (int64, error)

// deleteById(table *types.TableDefinition, id interface{}) error
// delete(table *types.TableDefinition, params map[string]string) (int64, error)

// insert(table *types.TableDefinition, attributes map[string]string) (int64, error)
//}

// type driver interface {
// 	session

// countBySingleTableInheritance(table *types.TableDefinition, params map[string]string) (int64, error)
// countInSpecificTable(table *types.TableDefinition, params map[string]string) (int64, error)

// findByIdAndSingleTableInheritance(table *types.TableDefinition, id interface{}, includes string) (map[string]interface{}, error)
// findByIdInSpecificTable(table *types.TableDefinition, id interface{}, includes string) (map[string]interface{}, error)
// findBySingleTableInheritance(table *types.TableDefinition, params map[string]string) ([]map[string]interface{}, error)
// findInSpecificTable(table *types.TableDefinition, params map[string]string) ([]map[string]interface{}, error)

// updateByIdAndSingleTableInheritance(table *types.TableDefinition, id interface{}, updated_attributes map[string]interface{}) (int64, error)
// updateByIdInSpecificTable(table *types.TableDefinition, id interface{}, updated_attributes map[string]interface{}) (int64, error)
// updateBySingleTableInheritance(table *types.TableDefinition, params map[string]string, updated_attributes map[string]interface{}) (int64, error)
// updateInSpecificTable(table *types.TableDefinition, params map[string]string, updated_attributes map[string]interface{}) (int64, error)

// deleteByIdAndSingleTableInheritance(table *types.TableDefinition, id interface{}) error
// deleteByIdInSpecificTable(table *types.TableDefinition, id interface{}) error
// deleteBySingleTableInheritance(table *types.TableDefinition, params map[string]string) (int64, error)
// deleteInSpecificTable(table *types.TableDefinition, params map[string]string) (int64, error)
//}

const (
	GENEERIC_DB = 0
	POSTGRESQL  = 1
	MSSQL       = 2
	ORACLE      = 3
	SQLITE      = 4
	MYSQL       = 5
)

func IsNumericParams(drv string) bool {
	switch drv {
	case "postgres":
		return true
	default:
		return false
	}
}
func GetDBType(drv string) int {
	switch drv {
	case "postgres":
		return POSTGRESQL
	case "mssql":
		return MSSQL
	case "sqlite":
		return SQLITE
	case "oracle":
		return ORACLE
	case "mysql":
		return MYSQL
	default:
		return GENEERIC_DB
	}
}

func replaceQuestion(buffer *bytes.Buffer, str string, idx int) (*bytes.Buffer, int) {
	s := []byte(str)
	for {
		i := bytes.IndexByte(s, '?')
		if -1 == i {
			buffer.Write(s)
			break
		}
		buffer.Write(s[:i])
		buffer.WriteString("$")
		buffer.WriteString(strconv.FormatInt(int64(idx), 10))
		s = s[i+1:]
		idx++
	}
	return buffer, idx
}

type driver struct {
	tables          *types.TableDefinitions
	drv             string
	dbType          int
	db              *sql.DB
	isNumericParams bool
	isOnly          bool
	//cti_policy      session
}

func newDriver(drvName string, db *sql.DB, tables *types.TableDefinitions) *driver {
	isOnly := *IsOnly
	dbType := GetDBType(drvName)

	if POSTGRESQL != dbType {
		isOnly = false
	}

	drv := &driver{tables: tables, drv: drvName, dbType: dbType, db: db,
		isNumericParams: IsNumericParams(drvName), isOnly: isOnly}
	//drv.cti_policy = &default_cti_policy{drv: drv}
	return drv
}

func (self *driver) newWhere(idx int,
	table *types.TableDefinition,
	buffer *bytes.Buffer) *whereBuilder {

	builder := &whereBuilder{tables: self.tables,
		table:     table,
		idx:       idx,
		isFirst:   true,
		prefix:    " WHERE ",
		buffer:    buffer,
		operators: default_operators,
		operators_for_field: map[string]map[string]op_func{"type": operators_for_type,
			"parent_type": operators_for_parent_type},
		add_argument: (*whereBuilder).appendNumericArguments}

	if self.isNumericParams {
		builder.add_argument = (*whereBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*whereBuilder).appendSimpleArguments
	}

	return builder
}

// ////////////////////////// count //////////////////////////
// func (self *driver) simpleCount(table *types.TableDefinition,
// 	params map[string]string, isSingleTableInheritance bool) (int64, error) {
// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT count(*) FROM ")
// 	if self.isOnly {
// 		buffer.WriteString("ONLY ")
// 	}
// 	buffer.WriteString(table.CollectionName)

// 	builder := self.newWhere(1, table, &buffer)

// 	if isSingleTableInheritance {
// 		builder.equalClass("type", table)
// 	}

// 	e := builder.build(params)
// 	if nil != e {
// 		return -1, e
// 	}

// 	row := self.db.QueryRow(buffer.String(), builder.params...)
// 	count := int64(0)
// 	e = row.Scan(&count)
// 	if nil != e {
// 		return 0, e
// 	}
// 	return count, nil
// }

// func (self *driver) countBySingleTableInheritance(table *types.TableDefinition,
// 	params map[string]string) (int64, error) {
// 	return self.simpleCount(table, params, true)
// }

// func (self *driver) countInSpecificTable(table *types.TableDefinition,
// 	params map[string]string) (int64, error) {
// 	return self.simpleCount(table, params, false)
// }

// func (self *driver) count(table *types.TableDefinition,
// 	params map[string]string) (int64, error) {
// 	if table.IsSingleTableInheritance() {
// 		return self.countBySingleTableInheritance(table, params)
// 	}

// 	if !table.HasChildren() {
// 		return self.countInSpecificTable(table, params)
// 	}

// 	return self.cti_policy.count(table, params)
// }

// ////////////////////////// query //////////////////////////
// func (self *driver) buildSQLQueryWithObjectId(table *types.TableDefinition) (QueryBuilder, error) {
// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT ")
// 	isSingleTableInheritance := table.IsSingleTableInheritance()
// 	columns := toColumns(table, isSingleTableInheritance)
// 	if nil == columns || 0 == len(columns) {
// 		return nil, errors.New("crazy! selected columns is empty.")
// 	}
// 	writeColumns(columns, &buffer)
// 	buffer.WriteString(" FROM ")
// 	if self.isOnly {
// 		buffer.WriteString("ONLY ")
// 	}
// 	buffer.WriteString(table.CollectionName)
// 	buffer.WriteString(" WHERE ")
// 	buffer.WriteString(table.Id.Name)
// 	if self.isNumericParams {
// 		buffer.WriteString(" = $1")
// 	} else {
// 		buffer.WriteString(" = ?")
// 	}

// 	return &QueryImpl{drv: self,
// 		table: table,
// 		isSingleTableInheritance: isSingleTableInheritance,
// 		columns:                  columns,
// 		sql:                      buffer.String()}, nil
// }
