package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Query interface {
	One() (map[string]interface{}, error)
	All() ([]map[string]interface{}, error)
}

type QueryBuilder interface {
	Bind(params ...interface{}) QueryBuilder
	Build() Query
}

func buildSelectStr(table *types.TableDefinition, buffer *bytes.Buffer) ([]*types.ColumnDefinition, error) {
	columns := make([]*types.ColumnDefinition, 0, len(table.GetAttributes()))
	isFirst := true
	for _, attribute := range table.GetAttributes() {
		if isFirst {
			isFirst = false
		} else {
			buffer.Write([]byte(", "))
		}

		buffer.Write([]byte(attribute.Name))
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
	columns, e := buildSelectStr(table, &buffer)
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

	return &QueryImpl{drv: drv, table: table, columns: columns, sql: buffer.String()}, nil
}

// func buildSQLWithSelectAndObjectId(drv *driver, table *types.TableDefinition,
// 	selectStr string, columns ...*types.ColumnDefinition) QueryBuilder {

// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT ")
// 	buffer.WriteString(selectStr)
// 	buffer.WriteString(" FROM ")
// 	buffer.WriteString(table.CollectionName)
// 	buffer.WriteString(" WHERE ")
// 	buffer.WriteString(table.Id.Name)
// 	if drv.isNumericParams {
// 		buffer.WriteString(" = $1")
// 	} else {
// 		buffer.WriteString(" = ?")
// 	}

// 	return &QueryImpl{drv: drv, table: table, columns: columns, sql: buffer.String()}
// }

// func buildQueryWithParams(drv *driver, table *types.TableDefinition,
// 	params map[string]string) (Query, error) {

// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT ")
// 	columns, e := buildSelectStr(table, &buffer)
// 	if nil != e {
// 		return nil, e
// 	}

// 	buffer.WriteString(" FROM ")
// 	buffer.WriteString(table.CollectionName)

// 	if nil == params {
// 		return &QueryImpl{drv: drv, table: table, columns: columns, sql: buffer.String()}, nil
// 	}

// 	args, e := buildWhereWithParams(drv, table, 1, params, &buffer)
// 	if nil != e {
// 		return nil, e
// 	}

// 	return &QueryImpl{drv: drv, table: table, columns: columns, sql: buffer.String(), parameters: args}, nil
// }

// func buildSQLWithSelectAndParams(drv *driver,
// 	table *types.TableDefinition,
// 	selectStr string,
// 	columns []*types.ColumnDefinition,
// 	params map[string]string) (Query, error) {

// 	var buffer bytes.Buffer
// 	buffer.WriteString("SELECT ")
// 	buffer.WriteString(selectStr)
// 	buffer.WriteString(" FROM ")
// 	buffer.WriteString(table.CollectionName)

// 	if nil == params {
// 		return &QueryImpl{drv: drv, table: table, columns: columns, sql: buffer.String()}, nil
// 	}

// 	args, e := buildWhereWithParams(drv, table, 1, params, &buffer)
// 	if nil != e {
// 		return nil, e
// 	}
// 	return &QueryImpl{drv: drv, table: table, columns: columns, sql: buffer.String(), parameters: args}, nil
// }

func whereWithParams(drv *driver, table *types.TableDefinition, isSimpleTableInheritance bool,
	idx int, params map[string]string, buffer *bytes.Buffer) ([]interface{}, error) {
	if nil == params {
		if !isSimpleTableInheritance {
			return nil, nil
		}
	}

	builder := whereBuilder{table: table,
		idx:          idx,
		isFirst:      true,
		prefix:       " WHERE ",
		buffer:       buffer,
		operators:    default_operators,
		add_argument: (*whereBuilder).appendNumericArguments}

	if drv.isNumericParams {
		builder.add_argument = (*whereBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*whereBuilder).appendSimpleArguments
	}

	if isSimpleTableInheritance {
		builder.equalClass("type", table)
	}

	if nil == params {
		return nil, nil
	}

	e := builder.build(params)
	if nil != e {
		return nil, e
	}

	if groupBy, ok := params["groupBy"]; ok {
		if 0 != len(groupBy) {
			return nil, errors.New("groupBy is empty.")
		}

		buffer.WriteString(" GROUP BY ")
		buffer.WriteString(groupBy)
	}

	if having, ok := params["having"]; ok {
		if 0 != len(having) {
			return nil, errors.New("having is empty.")
		}

		buffer.WriteString(" HAVING ")
		buffer.WriteString(having)
	}

	if order, ok := params["order"]; ok {
		if 0 != len(order) {
			return nil, errors.New("order is empty.")
		}

		buffer.WriteString(" ORDER BY ")
		buffer.WriteString(order)
	}

	if limit, ok := params["limit"]; ok {
		i, e := strconv.ParseInt(limit, 10, 64)
		if nil != e {
			return nil, fmt.Errorf("limit is not a number, actual value is '" + limit + "'")
		}
		if i <= 0 {
			return nil, fmt.Errorf("limit must is geater zero, actual value is '" + limit + "'")
		}

		if offset, ok := params["offset"]; ok {
			i, e = strconv.ParseInt(offset, 10, 64)
			if nil != e {
				return nil, fmt.Errorf("offset is not a number, actual value is '" + offset + "'")
			}

			if i < 0 {
				return nil, fmt.Errorf("offset must is geater(or equals) zero, actual value is '" + offset + "'")
			}

			buffer.WriteString(" LIMIT ")
			buffer.WriteString(offset)
			buffer.WriteString(" , ")
			buffer.WriteString(limit)
		} else {
			buffer.WriteString(" LIMIT ")
			buffer.WriteString(limit)
		}
	}

	return builder.params, nil
}

type QueryImpl struct {
	drv                      *driver
	columns                  []*types.ColumnDefinition
	sql                      string
	parameters               []interface{}
	isSimpleTableInheritance bool
	table                    *types.TableDefinition
}

func (self *QueryImpl) Bind(params ...interface{}) QueryBuilder {
	self.parameters = params
	return self
}

func (self *QueryImpl) Build() Query {
	return self
}

func (q *QueryImpl) One() (map[string]interface{}, error) {
	row := q.drv.db.QueryRow(q.sql, q.parameters...)

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

func collectDefinitions(table *types.TableDefinition,
	tables map[string]*types.TableDefinition,
	columnDefs map[string]*types.ColumnDefinition) {

	for _, child := range table.OwnChildren.All() {
		if nil != table.OwnAttributes {
			for k, column := range table.OwnAttributes {
				columnDefs[k] = column
			}
		}

		tables[child.UnderscoreName] = child
		if child.HasChildren() {
			collectDefinitions(child, tables, columnDefs)
		}
	}
}

func (q *QueryImpl) bySimpleTableInheritance(rows *sql.Rows) ([]map[string]interface{}, error) {
	columnNames, e := rows.Columns()
	if nil != e {
		return nil, e
	}
	columnDefinitions := map[string]*types.ColumnDefinition{}
	tables := map[string]*types.TableDefinition{}

	tables[q.table.UnderscoreName] = q.table
	if nil != q.table.Attributes {
		for k, column := range q.table.Attributes {
			columnDefinitions[k] = column
		}
	}
	if q.table.HasChildren() {
		collectDefinitions(q.table, tables, columnDefinitions)
	}

	columns := make([]*types.ColumnDefinition, 0, len(columnNames))
	typeIdx := -1
	for _, name := range columnNames {
		column := columnDefinitions[name]
		if nil == column {
			columns = append(columns, nil)
			continue
		}
		if "type" == name {
			typeIdx = len(columns)
		}
		columns = append(columns, column)
	}
	if -1 == typeIdx {
		return nil, errors.New("column 'type' is not found in the sql result.")
	}

	results := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		var scanResultContainer []interface{}
		for _, column := range columns {
			if nil == column {
				var value interface{}
				scanResultContainer = append(scanResultContainer, &value)
			} else {
				scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
			}
		}

		if e = rows.Scan(scanResultContainer...); nil != e {
			return nil, e
		}

		typeValue, e := toInternalValue(columns[typeIdx], scanResultContainer[typeIdx])
		if nil != e {
			return nil, fmt.Errorf("convert column 'type' to internal value failed, %v, value is [%T]%v",
				e, scanResultContainer[typeIdx], scanResultContainer[typeIdx])
		}
		t, ok := typeValue.(string)
		if !ok {
			return nil, errors.New("column 'type' is not a string")
		}
		table, ok := tables[t]
		if !ok {
			return nil, errors.New("table '" + t + "' is undefined")
		}

		res := map[string]interface{}{}
		for i, column := range columns {
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

		results = append(results, res)
	}

	if nil != rows.Err() {
		return nil, rows.Err()
	}
	return results, nil
}

func (q *QueryImpl) byColumns(rows *sql.Rows) ([]map[string]interface{}, error) {

	results := make([]map[string]interface{}, 0, 10)
	for rows.Next() {

		var scanResultContainer []interface{}
		for _, column := range q.columns {
			scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
		}

		if e := rows.Scan(scanResultContainer...); nil != e {
			return nil, e
		}

		res := map[string]interface{}{}
		for i, column := range q.columns {
			var e error
			res[column.Name], e = toInternalValue(column, scanResultContainer[i])
			if nil != e {
				return nil, fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v",
					column.Name, e, scanResultContainer[i], scanResultContainer[i])
			}
		}

		results = append(results, res)
	}

	if nil != rows.Err() {
		return nil, rows.Err()
	}
	return results, nil
}

func (q *QueryImpl) All() ([]map[string]interface{}, error) {
	fmt.Println(q.table.Name, q.sql)
	rs, e := q.drv.db.Prepare(q.sql)
	if e != nil {
		return nil, e
	}
	defer rs.Close()

	rows, e := rs.Query(q.parameters...)
	if e != nil {
		return nil, e
	}

	if q.isSimpleTableInheritance {
		return q.bySimpleTableInheritance(rows)
	} else {
		return q.byColumns(rows)
	}
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
