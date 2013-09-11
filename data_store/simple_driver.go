package data_store

import (
	"bytes"
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type driver interface {
	insert(table *types.TableDefinition, attributes map[string]interface{}) (int64, error)

	findById(table *types.TableDefinition, id interface{}) (map[string]interface{}, error)

	find(table *types.TableDefinition, params map[string]string) ([]map[string]interface{}, error)

	updateById(table *types.TableDefinition, id interface{},
		updated_attributes map[string]interface{}) error

	update(table *types.TableDefinition, params map[string]string,
		updated_attributes map[string]interface{}) (int64, error)

	deleteById(table *types.TableDefinition, id interface{}) error

	delete(table *types.TableDefinition, params map[string]string) (int64, error)

	count(table *types.TableDefinition, params map[string]string) (int64, error)

	snapshot(table *types.TableDefinition, params map[string]string) ([]map[string]interface{}, error)

	forEach(table *types.TableDefinition, params map[string]string,
		cb func(table *types.TableDefinition, id interface{}) error) error
}

var (
	id_column = &types.ColumnDefinition{AttributeDefinition: types.AttributeDefinition{Name: "id",
		Type:       types.GetTypeDefinition("objectId"),
		Collection: types.COLLECTION_UNKNOWN}}

	created_at_column = &types.ColumnDefinition{AttributeDefinition: types.AttributeDefinition{Name: "created_at",
		Type:       types.GetTypeDefinition("datetime"),
		Collection: types.COLLECTION_UNKNOWN}}

	updated_at_column = &types.ColumnDefinition{AttributeDefinition: types.AttributeDefinition{Name: "updated_at",
		Type:       types.GetTypeDefinition("datetime"),
		Collection: types.COLLECTION_UNKNOWN}}

	snapshot_columns = []*types.ColumnDefinition{id_column, created_at_column, updated_at_column}

	snapshot_definition = &types.TableDefinition{Name: "snapshot",
		UnderscoreName: "snapshot",
		CollectionName: "snapshot",
		Id:             id_column}
)

func init() {
	attributes := map[string]*types.ColumnDefinition{id_column.Name: id_column,
		created_at_column.Name: created_at_column,
		updated_at_column.Name: updated_at_column}

	snapshot_definition.OwnAttributes = attributes
	snapshot_definition.Attributes = attributes
}

const (
	GENEERIC_DB = 0
	POSTGRESQL  = 1
	MSSQL       = 2
	ORACLE      = 3
	SQLITE      = 4
	MYSQL       = 5
)

func IsNumericParams(drv int) bool {
	switch drv {
	case POSTGRESQL, ORACLE:
		return true
	default:
		return false
	}
}
func GetDBType(drv string) int {
	switch drv {
	case "postgres":
		return POSTGRESQL
	case "mssql", "sqlerver":
		return MSSQL
	case "sqlite":
		return SQLITE
	case "oracle":
		return ORACLE
	case "mysql", "mymysql":
		return MYSQL
	default:
		if strings.HasPrefix(drv, "odbc_with_") {
			switch drv[len("odbc_with_"):] {
			case "postgres":
				return POSTGRESQL
			case "mssql", "sqlerver":
				return MSSQL
			case "sqlite":
				return SQLITE
			case "oracle":
				return ORACLE
			case "mysql", "mymysql":
				return MYSQL
			}
		}
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

type simple_driver struct {
	tables *types.TableDefinitions
	drv    string
	dbType int
	db     *sql.DB

	isNumericParams bool
	hasOnly         bool
	from            string
}

func simpleDriver(drvName string, db *sql.DB, hasOnly bool, tables *types.TableDefinitions) *simple_driver {
	dbType := GetDBType(drvName)
	from := " FROM "
	if dbType == POSTGRESQL && hasOnly {
		from = " FROM ONLY "
	}
	return &simple_driver{tables: tables, drv: drvName, dbType: dbType, db: db,
		isNumericParams: IsNumericParams(dbType), from: from, hasOnly: hasOnly}
}

func (self *simple_driver) newWhere(idx int,
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

	switch self.dbType {
	case POSTGRESQL:
		builder.limit_and_offset = (*whereBuilder).limit_and_offset_postgres
	default:
		builder.limit_and_offset = (*whereBuilder).limit_and_offset_generic
	}

	return builder
}

func scanOne(row resultScan, columns []*types.ColumnDefinition) ([]interface{}, error) {
	var scanResultContainer []interface{}
	for _, column := range columns {
		scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
	}

	if e := row.Scan(scanResultContainer...); nil != e {
		return nil, e
	}

	res := make([]interface{}, 0, len(columns))
	for i, column := range columns {
		v, e := toInternalValue(column, scanResultContainer[i])
		if nil != e {
			if e == types.InvalidValueError {
				continue
			}

			return nil, fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v",
				column.Name, e, scanResultContainer[i], scanResultContainer[i])
		}
		res = append(res, v)
	}
	return res, nil
}

func (self *simple_driver) selectOne(sql string, args []interface{},
	columns []*types.ColumnDefinition) ([]interface{}, error) {
	row := self.db.QueryRow(sql, args...)
	return scanOne(row, columns)
}

func (self *simple_driver) selectAll(sql string, args []interface{},
	columns []*types.ColumnDefinition) ([][]interface{}, error) {
	rs, e := self.db.Prepare(sql)
	if e != nil {
		return nil, e
	}
	defer rs.Close()

	rows, e := rs.Query(args...)
	if e != nil {
		return nil, e
	}

	results := make([][]interface{}, 0, 10)
	for rows.Next() {
		res, e := scanOne(rows, columns)
		if e != nil {
			return nil, e
		}
		results = append(results, res)
	}

	if nil != rows.Err() {
		return nil, rows.Err()
	}
	return results, nil
}

func (self *simple_driver) count(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT count(*) ")
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)

	builder := self.newWhere(1, table, &buffer)

	if table.IsSingleTableInheritance() {
		builder.equalClass("type", table)
	}

	e := builder.build(params)
	if nil != e {
		return -1, e
	}

	row := self.db.QueryRow(buffer.String(), builder.params...)
	count := int64(0)
	e = row.Scan(&count)
	if nil != e {
		return 0, e
	}
	return count, nil
}

func (self *simple_driver) snapshot(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	writeColumns(snapshot_columns, &buffer)
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)
	builder := self.newWhere(1, table, &buffer)

	if table.IsSingleTableInheritance() {
		builder.equalClass("type", table)
	}

	e := builder.build(params)
	if nil != e {
		return nil, e
	}
	//fmt.Println(buffer.String(), builder.params)
	q := &QueryImpl{drv: self,
		isSingleTableInheritance: false,
		columns:                  snapshot_columns,
		table:                    snapshot_definition,
		sql:                      buffer.String(),
		parameters:               builder.params}

	return q.All()
}

func (self *simple_driver) buildSQLQueryWithObjectId(table *types.TableDefinition) (QueryBuilder, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	isSingleTableInheritance := table.IsSingleTableInheritance()
	columns := toColumns(table, isSingleTableInheritance)
	if nil == columns || 0 == len(columns) {
		return nil, errors.New("crazy! selected columns is empty.")
	}
	writeColumns(columns, &buffer)
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)
	buffer.WriteString(" WHERE ")
	buffer.WriteString(table.Id.Name)
	if self.isNumericParams {
		buffer.WriteString(" = $1")
	} else {
		buffer.WriteString(" = ?")
	}

	return &QueryImpl{drv: self,
		table: table,
		isSingleTableInheritance: isSingleTableInheritance,
		columns:                  columns,
		sql:                      buffer.String()}, nil
}

func (self *simple_driver) findById(table *types.TableDefinition,
	id interface{}) (map[string]interface{}, error) {
	builder, e := self.buildSQLQueryWithObjectId(table)
	if nil != e {
		return nil, e
	}
	return builder.Bind(id).Build().One()
}

func (self *simple_driver) where(table *types.TableDefinition, sqlString string) (QueryBuilder, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	isSingleTableInheritance := table.IsSingleTableInheritance()
	columns := toColumns(table, isSingleTableInheritance)
	if nil == columns || 0 == len(columns) {
		return nil, errors.New("crazy! selected columns is empty.")
	}
	writeColumns(columns, &buffer)
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)
	if 0 != len(sqlString) {
		buffer.WriteString(" WHERE ")
		replaceQuestion(&buffer, sqlString, 1)
	}

	return &QueryImpl{drv: self,
		isSingleTableInheritance: isSingleTableInheritance,
		columns:                  columns,
		table:                    table,
		sql:                      buffer.String()}, nil
}

func (self *simple_driver) find(table *types.TableDefinition,
	params map[string]string) ([]map[string]interface{}, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	isSingleTableInheritance := table.IsSingleTableInheritance()
	columns := toColumns(table, isSingleTableInheritance)
	if nil == columns || 0 == len(columns) {
		return nil, errors.New("crazy! selected columns is empty.")
	}
	writeColumns(columns, &buffer)
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)
	args, e := self.whereWithParams(table, isSingleTableInheritance, 1, params, &buffer)
	if nil != e {
		return nil, e
	}

	q := &QueryImpl{drv: self,
		isSingleTableInheritance: isSingleTableInheritance,
		columns:                  columns,
		table:                    table,
		sql:                      buffer.String(),
		parameters:               args}

	return q.All()
}

func (self *simple_driver) whereWithParams(table *types.TableDefinition, isSingleTableInheritance bool,
	idx int, params map[string]string, buffer *bytes.Buffer) ([]interface{}, error) {
	if nil == params {
		if !isSingleTableInheritance {
			return nil, nil
		}
	}

	builder := self.newWhere(idx, table, buffer)

	if isSingleTableInheritance {
		builder.equalClass("type", table)
	}

	if nil == params {
		return nil, nil
	}

	e := builder.buildSQL(params)
	if nil != e {
		return nil, e
	}

	return builder.params, nil
}

////////////////////////// insert //////////////////////////
func typeFrom(table *types.TableDefinition, attributes map[string]interface{}) (*types.TableDefinition, error) {
	t, ok := attributes["type"]
	if !ok {
		return table, nil
	}

	nm, ok := t.(string)
	if !ok {
		return nil, fmt.Errorf("'type' must is a string, actual type is a %T, actual value is '%v'.", t, t)
	}

	if nm == table.UnderscoreName {
		return table, nil
	}

	defintion := table.FindByUnderscoreName(nm)
	if nil == defintion {
		return nil, errors.New("table '" + nm + "' with parent was table '" + table.UnderscoreName + "' is not exists.")
	}

	if !defintion.IsSubclassOf(table) {
		return nil, errors.New("table '" + nm + "' is not inherit from table '" + table.UnderscoreName + "'")
	}

	return defintion, nil
}

func (self *simple_driver) insert(table *types.TableDefinition,
	attributes map[string]interface{}) (int64, error) {
	var e error
	table, e = typeFrom(table, attributes)
	if nil != e {
		return 0, e
	}

	if table.IsAbstract {
		return 0, errors.New("table '" + table.Name + "' is abstract.")
	}

	var buffer bytes.Buffer
	var values bytes.Buffer
	params := make([]interface{}, 0, len(table.Attributes))

	idx := 1
	for _, attribute := range table.Attributes {
		//////////////////////////////////////////
		// TODO: refactor it?
		if attribute.IsSerial() {
			continue
		}

		var value interface{} = nil
		switch attribute.Name {
		case "created_at":
			fallthrough
		case "updated_at":
			value = time.Now()
		case "type":
			value = table.UnderscoreName
		default:
			v := attributes[attribute.Name]
			if nil == v {
				if attribute.IsRequired {
					return 0, fmt.Errorf("column '%v' is required", attribute.Name)
				}
				if nil == attribute.DefaultValue {
					continue
				}
				v = attribute.DefaultValue
			}

			value, e = attribute.Type.ToInternal(v)
			if nil != e {
				if e == types.InvalidValueError {
					continue
				}
				return 0, fmt.Errorf("column '%v' is not a '%v', actual value is '%v', %v",
					attribute.Name, attribute.Type.Name(), v, e)
			}
			value = attribute.Type.ToExternal(value)
		}
		//////////////////////////////////////////

		if 1 != idx {
			buffer.WriteString(", ")
			values.WriteString(", ")
		}
		buffer.WriteString(attribute.Name)
		if self.isNumericParams {
			values.WriteString("$")
			values.WriteString(strconv.FormatInt(int64(idx), 10))
		} else {
			values.WriteString("?")
		}
		params = append(params, value)

		idx++
	}

	sql := "INSERT INTO " + table.CollectionName + "( " + buffer.String() +
		" ) VALUES ( " + values.String() + " )"

	var id int64
	if "postgres" == self.drv {
		e = self.db.QueryRow(sql+" RETURNING "+table.Id.Name, params...).Scan(&id)
		if nil != e {
			return 0, e
		}
	} else {
		res, e := self.db.Exec(sql, params...)
		if nil != e {
			return 0, e
		}

		if affected, e := res.RowsAffected(); 1 != affected {
			return 0, fmt.Errorf("insert to %v failed, affected rows is %v %v",
				table.CollectionName, affected, e)
		}

		id, e = res.LastInsertId()
		if nil != e {
			return 0, e
		}
	}

	return id, nil
}

func (self *simple_driver) update(table *types.TableDefinition, params map[string]string,
	updated_attributes map[string]interface{}) (int64, error) {
	var buffer bytes.Buffer
	builder := updateBuilder{drv: self,
		table:  table,
		buffer: &buffer,
		idx:    1}
	e := builder.buildUpdate(updated_attributes)
	if nil != e {
		return 0, e
	}

	if nil == builder.params || 0 == len(builder.params) {
		return 0, errors.New("updated attributes is empty.")
	}

	e = builder.buildWhere(params, table.IsSingleTableInheritance())
	if nil != e {
		return 0, e
	}

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *simple_driver) updateBySQL(table *types.TableDefinition,
	updated_attributes map[string]interface{},
	queryString string, args ...interface{}) (int64, error) {
	var buffer bytes.Buffer
	builder := updateBuilder{drv: self,
		table:  table,
		buffer: &buffer,
		idx:    1}
	e := builder.buildUpdate(updated_attributes)
	if nil != e {
		return 0, e
	}

	if nil == builder.params || 0 == len(builder.params) {
		return 0, errors.New("updated attributes is empty.")
	}

	builder.buildWhereWithString(queryString, args)

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *simple_driver) updateById(table *types.TableDefinition, id interface{},
	updated_attributes map[string]interface{}) error {
	var buffer bytes.Buffer
	builder := updateBuilder{drv: self,
		table:  table,
		buffer: &buffer,
		idx:    1}
	e := builder.buildUpdate(updated_attributes)
	if nil != e {
		return e
	}

	if nil == builder.params || 0 == len(builder.params) {
		return errors.New("updated attributes is empty.")
	}

	builder.buildWhereById(id)

	//fmt.Println(buffer.String(), builder.params)
	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return e
	}

	affected, e := res.RowsAffected()
	if nil != e {
		return e
	}

	switch affected {
	case 0:
		return sql.ErrNoRows
	case 1:
		return nil
	default:
		return fmt.Errorf("affected rows is not equals 1, actual is %v", affected)
	}
}

func (self *simple_driver) delete(table *types.TableDefinition,
	params map[string]string) (int64, error) {
	var buffer bytes.Buffer
	buffer.WriteString("DELETE ")
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)

	builder := self.newWhere(1, table, &buffer)
	if table.IsSingleTableInheritance() {
		builder.equalClass("type", table)
	}

	e := builder.build(params)
	if nil != e {
		return 0, e
	}

	res, e := self.db.Exec(buffer.String(), builder.params...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *simple_driver) deleteById(table *types.TableDefinition, id interface{}) error {
	var buffer bytes.Buffer
	buffer.WriteString("DELETE ")
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)
	buffer.WriteString(" WHERE ")
	buffer.WriteString(table.Id.Name)
	if self.isNumericParams {
		buffer.WriteString(" = $1")
	} else {
		buffer.WriteString(" = ?")
	}

	res, e := self.db.Exec(buffer.String(), table.Id.Type.ToExternal(id))
	if nil != e {
		return e
	}

	affected, e := res.RowsAffected()
	if nil != e {
		return e
	}

	switch affected {
	case 0:
		return sql.ErrNoRows
	case 1:
		return nil
	default:
		return fmt.Errorf("affected rows is not equals 1, actual is %v", affected)
	}
}

func (self *simple_driver) deleteBySQL(table *types.TableDefinition,
	queryString string, args ...interface{}) (int64, error) {

	var buffer bytes.Buffer
	buffer.WriteString("DELETE ")
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)

	if 0 == len(queryString) {
		buffer.WriteString(" WHERE ")
		if self.isNumericParams {
			_, c := replaceQuestion(&buffer, queryString, 1)
			if len(args) != c {
				return 0, errors.New("parameters count is error")
			}
		} else {
			buffer.WriteString(queryString)
		}
	}

	res, e := self.db.Exec(buffer.String(), args...)
	if nil != e {
		return 0, e
	}

	return res.RowsAffected()
}

func (self *simple_driver) forEach(table *types.TableDefinition, params map[string]string,
	cb func(table *types.TableDefinition, id interface{}) error) error {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	buffer.WriteString(table.Id.Name)
	buffer.WriteString(self.from)
	buffer.WriteString(table.CollectionName)

	builder := self.newWhere(1, table, &buffer)
	if table.IsSingleTableInheritance() {
		builder.equalClass("type", table)
	}

	e := builder.build(params)
	if nil != e {
		return e
	}

	rs, e := self.db.Prepare(buffer.String())
	if e != nil {
		return e
	}
	defer rs.Close()

	rows, e := rs.Query(builder.params...)
	if e != nil {
		return e
	}

	for rows.Next() {
		id_value := table.Id.Type.MakeValue()

		if e := rows.Scan(id_value); nil != e {
			return e
		}

		id, e := toInternalValue(table.Id, id_value)
		if nil != e {
			if e == types.InvalidValueError {
				continue
			}

			return fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v",
				table.Id.Name, e, id_value, id_value)
		}
		e = cb(table, id)
		if nil != e {
			return e
		}
	}

	if nil != rows.Err() {
		return rows.Err()
	}
	return nil
}
