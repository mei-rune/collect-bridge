package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
)

type Query interface {
	One() (map[string]interface{}, error)
	All() ([]map[string]interface{}, error)
}

type QueryBuilder interface {
	Bind(params ...interface{}) QueryBuilder
	Build() Query
}

func collectColumns(table *types.TableDefinition,
	columnDefs map[string]*types.ColumnDefinition) {
	for _, child := range table.OwnChildren.All() {
		if nil != table.OwnAttributes {
			for k, column := range table.OwnAttributes {
				columnDefs[k] = column
			}
		}

		if child.HasChildren() {
			collectColumns(child, columnDefs)
		}
	}
}

func mergeColumns(table *types.TableDefinition) map[string]*types.ColumnDefinition {
	columns := make(map[string]*types.ColumnDefinition)
	if nil != table.Attributes {
		for k, column := range table.Attributes {
			columns[k] = column
		}
	}

	if !table.HasChildren() {
		return columns
	}

	collectColumns(table, columns)
	return columns
}

func buildSelectStr(table *types.TableDefinition, isSingleTableInheritance bool,
	buffer *bytes.Buffer) ([]*types.ColumnDefinition, error) {
	columns := make([]*types.ColumnDefinition, 0, len(table.GetAttributes()))
	isFirst := true
	var attributes map[string]*types.ColumnDefinition

	if isSingleTableInheritance {
		attributes = mergeColumns(table)
		attribute, ok := attributes["type"]
		if !ok {
			panic("table '" + table.Name + "' is simple table inheritance, but it is not contains column 'type'.")
		}

		delete(attributes, "type")
		buffer.WriteString(attribute.Name)
		columns = append(columns, attribute)
		isFirst = false
	} else {
		attributes = table.GetAttributes()
	}

	for _, attribute := range attributes {
		if isFirst {
			isFirst = false
		} else {
			buffer.Write([]byte(", "))
		}

		buffer.WriteString(attribute.Name)
		columns = append(columns, attribute)
	}

	if nil == columns || 0 == len(columns) {
		return nil, errors.New("crazy! selected columns is empty.")
	}
	return columns, nil
}

func buildSQLQueryWithObjectId(drv *driver, table *types.TableDefinition) (QueryBuilder, error) {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	isSingleTableInheritance := table.IsSingleTableInheritance()
	columns, e := buildSelectStr(table, isSingleTableInheritance, &buffer)
	if nil != e {
		return nil, e
	}
	buffer.WriteString(" FROM ")
	buffer.WriteString(table.CollectionName)
	buffer.WriteString(" WHERE ")
	buffer.WriteString(table.Id.Name)
	if drv.isNumericParams {
		buffer.WriteString(" = $1")
	} else {
		buffer.WriteString(" = ?")
	}
	return &QueryImpl{drv: drv,
		table: table,
		isSingleTableInheritance: isSingleTableInheritance,
		columns:                  columns,
		sql:                      buffer.String()}, nil
}

type QueryImpl struct {
	drv                      *driver
	columns                  []*types.ColumnDefinition
	sql                      string
	parameters               []interface{}
	isSingleTableInheritance bool
	table                    *types.TableDefinition
}

func (self *QueryImpl) Bind(params ...interface{}) QueryBuilder {
	self.parameters = params
	return self
}

func (self *QueryImpl) Build() Query {
	return self
}

type resultScan interface {
	Scan(dest ...interface{}) error
}

func (q *QueryImpl) rowbySingleTableInheritance(rows resultScan) (map[string]interface{}, error) {
	var scanResultContainer []interface{}
	for _, column := range q.columns {
		if nil == column {
			var value interface{}
			scanResultContainer = append(scanResultContainer, &value)
		} else {
			scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
		}
	}

	if e := rows.Scan(scanResultContainer...); nil != e {
		return nil, e
	}

	typeValue, e := toInternalValue(q.columns[0], scanResultContainer[0])
	if nil != e {
		return nil, fmt.Errorf("convert column 'type' to internal value failed, %v, value is [%T]%v",
			e, scanResultContainer[0], scanResultContainer[0])
	}
	t, ok := typeValue.(string)
	if !ok {
		return nil, errors.New("column 'type' is not a string")
	}
	table := q.table.FindByUnderscoreName(t)
	if nil == table {
		return nil, errors.New("table '" + t + "' is undefined")
	}

	res := map[string]interface{}{}
	res["type"] = t
	for i, column := range q.columns {
		if nil == column {
			continue
		}

		if hasColumn := table.GetAttribute(column.Name); nil == hasColumn {
			continue
		}

		res[column.Name], e = toInternalValue(column, scanResultContainer[i])
		if nil != e {
			return nil, fmt.Errorf("convert column '%v' to internal value failed, %v, value is [%T]%v",
				column.Name, e, scanResultContainer[i], scanResultContainer[i])
		}
	}
	return res, nil
}

func (q *QueryImpl) bySingleTableInheritance(rows *sql.Rows) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		res, e := q.rowbySingleTableInheritance(rows)
		if nil != e {
			return nil, e
		}
		results = append(results, res)
	}

	if nil != rows.Err() {
		return nil, rows.Err()
	}
	return results, nil
}

func (q *QueryImpl) rowbyColumns(row resultScan) (map[string]interface{}, error) {
	var scanResultContainer []interface{}
	for _, column := range q.columns {
		scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
	}

	e := row.Scan(scanResultContainer...)
	if nil != e {
		return nil, e
	}

	res := make(map[string]interface{})
	for i, column := range q.columns {
		res[column.Name], e = toInternalValue(column, scanResultContainer[i])
		if nil != e {
			return nil, fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v",
				column.Name, e, scanResultContainer[i], scanResultContainer[i])
		}
	}
	return res, nil
}

func (q *QueryImpl) byColumns(rows *sql.Rows) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		res, e := q.rowbyColumns(rows)
		if nil != e {
			return nil, e
		}
		results = append(results, res)
	}

	if nil != rows.Err() {
		return nil, rows.Err()
	}
	return results, nil
}

func (q *QueryImpl) One() (map[string]interface{}, error) {
	row := q.drv.db.QueryRow(q.sql, q.parameters...)
	if q.isSingleTableInheritance {
		return q.rowbySingleTableInheritance(row)
	} else {
		return q.rowbyColumns(row)
	}
}

func (q *QueryImpl) All() ([]map[string]interface{}, error) {
	rs, e := q.drv.db.Prepare(q.sql)
	if e != nil {
		return nil, e
	}
	defer rs.Close()

	rows, e := rs.Query(q.parameters...)
	if e != nil {
		return nil, e
	}

	if q.isSingleTableInheritance {
		return q.bySingleTableInheritance(rows)
	} else {
		return q.byColumns(rows)
	}
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
			return nil, fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v",
				column.Name, e, scanResultContainer[i], scanResultContainer[i])
		}
		res = append(res, v)
	}
	return res, nil
}

func selectOne(drv *driver, sql string, args []interface{},
	columns ...*types.ColumnDefinition) ([]interface{}, error) {
	row := drv.db.QueryRow(sql, args...)
	return scanOne(row, columns)
}

func selectAll(drv *driver, sql string, args []interface{},
	columns ...*types.ColumnDefinition) ([][]interface{}, error) {
	rs, e := drv.db.Prepare(sql)
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

func toInternalValue(column *types.ColumnDefinition, v interface{}) (interface{}, error) {
	return column.Type.ToInternal(v)
}

// type Query {
//   One()(map[string]interface{}, error)
//   Some()([]map[string]interface{}, error)
// }

// type Query1 {
//   columns
//   sql
//   params
// }

// func (self *Query1) Bind(params) QueryBuilder {
//   self.params = params
//   return self
// }

// func (self *Query1) Build() Query {
//   return self
// }

// func (self *Query1) One()(map[string]interface{}, error) {
// }

// func  (self *Query1) Some()([]map[string]interface{}, error) {
// }

// type Query2 {
//   tables
//   sql
//   params
// }

// func (self *Query2) Bind(params) QueryBuilder {
//   self.params = params
//   return self
// }

// func (self *Query2) Build() Query {
//   return self
// }

// func (self *Query2) One()(map[string]interface{}, error) {
// }

// func (self *Query2) Some()([]map[string]interface{}, error) {
//   for {
//     nm := getAttribute("type")
//     columns = self.tables[nm].toColumns()
//     self.scan(column)
//   }
// }

// type QueryBuilder {
//   Bind(params) Query
// }

// queryByParams(table, params) ([]map[string]interface{}, error) {
//    columns := table.toColumns()
//    sql, args := buildSqlByParams(columns, table, params)
//    return & Query1{columns: columns, table: table, sql, args) ([]map[string]interface{}, error)
// }

// stiqueryByParams(tables, table, params) ([]map[string]interface{}, error) {
//    sql, args := buildSqlWithSelectByParams("*", table, params)
//    return fetchSomeBySti(tables, sql, args) ([]map[string]interface{}, error)
// }

// queryById(table, id) (map[string]interface{}, error) {
//   columns := table.toColumns()
//   sql := buildSqlById(columns, table)
//   return fetchOne(columns, "", id)
// }

// queryByParamsFromMutiTable(tables, table, id) (map[string]interface{}, error) {
//    if !params.has(orderBy) {
//      params["order"] = "tableoid"
//    }
//    sql, args := buildSqlWithSelectByParams("tableoid::regclass as name, id as id", table, params)

//    results = nil
//    last_nm = nil
//    query = nil
//    for _, r := range fetchSome(columns, table, sql, args) ([]map[string]interface{}, error) {
//       nm = r["name"]
//       id := r["id"]

//       if nil == query || nm != last_nm {
//          query = buildQueryById(tables[nm])
//       }
//       results = append(results, query.Bind(id).Build().One())
//    }
//    return results, nil
// }

// buildSqlById(columns, table) QueryBuilder {
//   return &Query1 {
//       columns: columns
//       sql: genSql(),
//       params}
// }

// buildSqlWithSelectByParams(selectStr, table, params) QueryBuilder {
//   return &Query1 {columns: columns
//     sql: genSql(selectStr),
//     params: params}
// }

// buildSqlByParams(columns, table, params) {
//   return buildSqlWithSelectByParams(join(columns), table, params)
// }
