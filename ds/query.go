package ds

import (
	"bytes"
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Iter interface {
	Next(res map[string]interface{}) bool
	Err() error
	Close()
}

type Query interface {
	One() (map[string]interface{}, error)
	Limit(start int64, size ...int64) Query
	Offset(offset int64) Query
	OrderBy(order string) Query
	Select(colums ...string) Query
	GroupBy(keys string) Query
	Having(conditions string) Query
	Iter() (Iter, error)
}

type QueryImpl struct {
	drv        string
	db         *sql.DB
	table      *types.TableDefinition
	limit      int64
	offset     int64
	where      string
	parameters []interface{}
	order      string
	groupBy    string
	having     string

	columns []*types.ColumnDefinition
	err     error
}

func (orm *QueryImpl) Limit(start int64, size ...int64) Query {
	orm.limit = start
	if len(size) > 0 {
		orm.offset = size[0]
	}
	return orm
}

func (orm *QueryImpl) Offset(offset int64) Query {
	orm.offset = offset
	return orm
}

func (orm *QueryImpl) OrderBy(order string) Query {
	orm.order = order
	return orm
}

func (orm *QueryImpl) Select(columns ...string) Query {
	if 0 == len(columns) {
		panic("you are crazy, columns must is not empty.")
	}
	var missing []string
	for _, nm := range columns {
		column := orm.table.GetAttribute(nm)
		if nil == column {
			missing = append(missing, nm)
			continue
		}
		orm.columns = append(orm.columns, column)
	}
	if nil != missing && 0 != len(missing) {
		orm.err = errors.New("column '" + strings.Join(missing, ",") + "' is not exists.")
	}
	return orm
}

func (orm *QueryImpl) attributes() []*types.ColumnDefinition {
	if nil == orm.columns || 0 == len(orm.columns) {
		for _, column := range orm.table.Attributes {
			orm.columns = append(orm.columns, column)
		}
	}
	return orm.columns
}

func (orm *QueryImpl) GroupBy(keys string) Query {
	orm.groupBy = keys
	return orm
}

func (orm *QueryImpl) Having(conditions string) Query {
	orm.having = conditions
	return orm
}

func (orm *QueryImpl) generateSql() string {
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")

	columns := orm.attributes()
	switch len(columns) {
	case 0:
		panic("crazy! table '" + orm.table.CollectionName + "' is emtpy columns")
	case 1:
		buffer.WriteString(columns[0].Name)
	default:
		buffer.WriteString(columns[0].Name)
		for _, column := range columns[1:] {
			buffer.WriteString(", ")
			buffer.WriteString(column.Name)
		}
	}

	buffer.WriteString(" FROM ")
	buffer.WriteString(orm.table.CollectionName)
	if 0 != len(orm.where) {
		buffer.WriteString(" WHERE ")
		buffer.WriteString(orm.where)
	}

	if 0 != len(orm.groupBy) {
		buffer.WriteString(" GROUP BY ")
		buffer.WriteString(orm.groupBy)
	}

	if 0 != len(orm.having) {
		buffer.WriteString(" HAVING ")
		buffer.WriteString(orm.having)
	}
	if 0 != len(orm.order) {
		buffer.WriteString(" ORDER BY ")
		buffer.WriteString(orm.order)
	}
	if orm.offset > 0 {
		buffer.WriteString(" LIMIT ")
		buffer.WriteString(strconv.FormatInt(orm.offset, 10))
		buffer.WriteString(" , ")
		buffer.WriteString(strconv.FormatInt(orm.limit, 10))
	} else if orm.limit > 0 {
		buffer.WriteString(" LIMIT ")
		buffer.WriteString(strconv.FormatInt(orm.limit, 10))
	}
	return buffer.String()
}

func (orm *QueryImpl) One() (map[string]interface{}, error) {
	columns := orm.attributes()
	if nil != orm.err {
		return nil, orm.err
	}

	row := orm.db.QueryRow(orm.generateSql(), orm.parameters...)

	var scanResultContainer []interface{}
	for _, column := range columns {
		scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
	}

	e := row.Scan(scanResultContainer...)
	if nil != e {
		return nil, e
	}

	res := make(map[string]interface{})
	for i, column := range columns {
		res[column.Name], e = toInternalValue(column, scanResultContainer[i])
		if nil != e {
			return nil, fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v", column.Name, e, scanResultContainer[i], scanResultContainer[i])
		}
	}
	return res, nil
}

func (orm *QueryImpl) Iter() (Iter, error) {
	colums := orm.attributes()
	if nil != orm.err {
		return nil, orm.err
	}
	rs, err := orm.db.Prepare(orm.generateSql())
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	res, err := rs.Query(orm.parameters...)
	if err != nil {
		return nil, err
	}
	return &IterImpl{rows: res, columns: colums}, nil
}

type IterImpl struct {
	table   *types.TableDefinition
	columns []*types.ColumnDefinition
	rows    *sql.Rows
	err     error
}

func (self *IterImpl) Close() {
	self.rows.Close()
}

func (self *IterImpl) Err() error {
	return self.err
}

func (self *IterImpl) Next(res map[string]interface{}) bool {
	if !self.rows.Next() {
		return false
	}

	var scanResultContainer []interface{}
	for _, column := range self.columns {
		scanResultContainer = append(scanResultContainer, column.Type.MakeValue())
	}

	if self.err = self.rows.Scan(scanResultContainer...); nil != self.err {
		return false
	}

	for i, column := range self.columns {
		var e error
		res[column.Name], e = toInternalValue(column, scanResultContainer[i])
		if nil != e {
			self.err = fmt.Errorf("convert %v to internal value failed, %v, value is [%T]%v", column.Name, e, scanResultContainer[i], scanResultContainer[i])
			return false
		}
	}

	return true
}

func toInternalValue(column *types.ColumnDefinition, v interface{}) (interface{}, error) {
	return column.Type.ToInternal(v)
}
