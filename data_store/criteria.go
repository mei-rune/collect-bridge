package data_store

import (
	"bytes"
	"commons/types"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// func buildClassQuery(cls *ClassDefinition) interface{} {
// 	cm := stringutils.Underscore(cls.Name)
// 	if !cls.IsInheritance() {
// 		return cm
// 	}
// 	if nil == cls.Children || 0 == len(cls.Children) {
// 		return cm
// 	}

// 	ar := make([]interface{}, 0, len(cls.Children))
// 	ar = append(ar, cm)
// 	for _, child := range cls.Children {
// 		ar = append(ar, stringutils.Underscore(child.Name))
// 	}
// 	return bson.M{"$in": ar}
// }

// func (self *mdb_server) buildClassQueryFromClassName(t string) error {
// 	cls := self.definitions.FindByUnderscoreName(t)
// 	if nil == cls {
// 		return nil, errors.New("class '" + t + "' is not found")
// 	}
// 	return buildClassQuery(cls), nil
// }

type op_func func(self *whereBuilder, column *types.ColumnDefinition, v string) error

var (
	default_operators = make(map[string]op_func)

	operators_for_type        = make(map[string]op_func)
	operators_for_parent_type = make(map[string]op_func)
)

func init() {
	default_operators["exists"] = (*whereBuilder).exists
	default_operators["in"] = (*whereBuilder).in
	default_operators["nin"] = (*whereBuilder).nin
	default_operators[">"] = (*whereBuilder).gt
	default_operators["gt"] = (*whereBuilder).gt
	default_operators[">="] = (*whereBuilder).gte
	default_operators["gte"] = (*whereBuilder).gte
	default_operators["="] = (*whereBuilder).eq
	default_operators["eq"] = (*whereBuilder).eq
	default_operators["!="] = (*whereBuilder).ne
	default_operators["ne"] = (*whereBuilder).ne
	default_operators["<"] = (*whereBuilder).lt
	default_operators["lt"] = (*whereBuilder).lt
	default_operators["<="] = (*whereBuilder).lte
	default_operators["lte"] = (*whereBuilder).lte
	default_operators["between"] = (*whereBuilder).between
	default_operators["is"] = (*whereBuilder).is
	default_operators["like"] = (*whereBuilder).like

	operators_for_type["eq"] = (*whereBuilder).equals_class
	operators_for_type["in"] = (*whereBuilder).in_class

	operators_for_parent_type["eq"] = (*whereBuilder).equals_parentClass
	operators_for_parent_type["in"] = (*whereBuilder).in_parentClass
}

type whereBuilder struct {
	tables  *types.TableDefinitions
	table   *types.TableDefinition
	idx     int
	isFirst bool
	prefix  string

	buffer *bytes.Buffer
	params []interface{}

	operators_for_field map[string]map[string]op_func
	operators           map[string]op_func
	add_argument        func(self *whereBuilder)
	limit_and_offset    func(self *whereBuilder, limit, offset string)
}

func (self *whereBuilder) appendArguments() {
	if nil != self.add_argument {
		self.add_argument(self)
	} else {
		self.appendSimpleArguments()
	}
}

func (self *whereBuilder) appendNumericArguments() {
	self.buffer.WriteString(" $")
	self.buffer.WriteString(fmt.Sprint(self.idx))
	self.idx++
}

func (self *whereBuilder) appendSimpleArguments() {
	self.buffer.WriteString(" ?")
	self.idx++
}

func (self *whereBuilder) append(ss ...string) error {
	if 0 == len(ss) {
		return nil
	}
	if self.isFirst {
		self.isFirst = false
		if 0 != len(self.prefix) {
			self.buffer.WriteString(self.prefix)
		}
	} else {
		self.buffer.WriteString(" AND ")
	}
	for _, s := range ss {
		self.buffer.WriteString(" ")
		self.buffer.WriteString(s)
	}
	return nil
}

func (self *whereBuilder) limit_and_offset_postgres(limit, offset string) {
	if 0 != len(limit) {
		self.buffer.WriteString(" LIMIT ")
		self.buffer.WriteString(limit)
	}

	if 0 != len(offset) {
		self.buffer.WriteString(" OFFSET ")
		self.buffer.WriteString(offset)
	}
}

func (self *whereBuilder) limit_and_offset_generic(limit, offset string) {
	if len(offset) != 0 {
		self.buffer.WriteString(" LIMIT ")
		self.buffer.WriteString(offset)
		self.buffer.WriteString(" , ")
		self.buffer.WriteString(limit)
	} else {
		self.buffer.WriteString(" LIMIT ")
		self.buffer.WriteString(limit)
	}
}

func (self *whereBuilder) add(column *types.ColumnDefinition, op, s string) error {
	v, e := column.Type.Parse(s)
	if nil != e {
		return fmt.Errorf("column '%v' convert '%v' to '%v' failed, %v",
			column.Name, s, column.Type.Name(), e)
	}

	if self.isFirst {
		self.isFirst = false
		if 0 != len(self.prefix) {
			self.buffer.WriteString(self.prefix)
		}
		self.buffer.WriteString(" ")
	} else {
		self.buffer.WriteString(" AND ")
	}

	self.buffer.WriteString(column.Name)
	self.buffer.WriteString(" ")
	self.buffer.WriteString(op)
	self.appendArguments()

	self.params = append(self.params, column.Type.ToExternal(v))
	return nil
}

func (self *whereBuilder) exists(column *types.ColumnDefinition, s string) error {
	return errors.New("not implemented")
}

func (self *whereBuilder) in(column *types.ColumnDefinition, s string) error {
	switch column.Type.Name() {
	case "ipAddress", "physicalAddress", "string":
		ss := strings.Split(s, ",")
		return self.append(column.Name, "IN ( '"+strings.Join(ss, "', '")+"' )")
	case "objectId", "integer", "decimal":
		return self.append(column.Name, "IN (", s, ")")
	default:
		return fmt.Errorf("'in' is not supported for the column '%v' ", column.Name)
	}
}

func (self *whereBuilder) nin(column *types.ColumnDefinition, s string) error {
	switch column.Type.Name() {
	case "ipAddress", "physicalAddress", "string":
		ss := strings.Split(s, ",")
		return self.append(column.Name, "NOT IN ( '"+strings.Join(ss, "', '")+"' )")
	case "objectId", "integer", "decimal":
		return self.append(column.Name, "NOT IN (", s, ")")
	default:
		return fmt.Errorf("'in' is not supported for the column '%v' ", column.Name)
	}
}

func (self *whereBuilder) gt(column *types.ColumnDefinition, s string) error {
	return self.add(column, ">", s)
}

func (self *whereBuilder) gte(column *types.ColumnDefinition, s string) error {
	return self.add(column, ">=", s)
}

func (self *whereBuilder) eq(column *types.ColumnDefinition, s string) error {
	return self.add(column, "=", s)
}

func (self *whereBuilder) ne(column *types.ColumnDefinition, s string) error {
	return self.add(column, "!=", s)
}

func (self *whereBuilder) lt(column *types.ColumnDefinition, s string) error {
	return self.add(column, "<", s)
}

func (self *whereBuilder) lte(column *types.ColumnDefinition, s string) error {
	return self.add(column, "<=", s)
}

func (self *whereBuilder) is(column *types.ColumnDefinition, s string) error {
	switch s {
	case "null", "NULL":
		return self.append(column.Name, "IS NULL")
	case "notnull", "NOTNULL":
		return self.append(column.Name, "IS NOT NULL")
	case "true", "TRUE":
		return self.append(column.Name, "IS TRUE")
	case "false", "FALSE":
		return self.append(column.Name, "IS FALSE")
	default:
		return fmt.Errorf("'is' is not supported with value '%v' for the column '%v' ", s, column.Name)
	}
}

func (self *whereBuilder) like(column *types.ColumnDefinition, s string) error {
	if "string" == column.Type.Name() {
		return fmt.Errorf("'like' is not supported for the column '%v', it must is a string type", column.Name)
	}

	return self.append(column.Name, " LIKE '"+s+"'")
}

func (self *whereBuilder) between(column *types.ColumnDefinition, s string) error {
	i := strings.IndexRune(s, ',')
	if -1 == i {
		return errors.New("column '" + column.Name + "' syntex error, it must has two value - '" + s + "'")
	}

	v1, e := column.Type.Parse(s[:i])
	if nil != e {
		return fmt.Errorf("column '%v' convert '%v' to '%v' failed, %v",
			column.Name, s[:i], column.Type.Name(), e)
	}

	v2, e := column.Type.Parse(s[i+1:])
	if nil != e {
		return fmt.Errorf("column '%v' convert '%v' to '%v' failed, %v",
			column.Name, s[i+1:], column.Type.Name(), e)
	}

	if self.isFirst {
		self.isFirst = false
		if 0 != len(self.prefix) {
			self.buffer.WriteString(self.prefix)
		}
		self.buffer.WriteString(" (")
	} else {
		self.buffer.WriteString(" AND (")
	}

	self.buffer.WriteString(column.Name)
	self.buffer.WriteString(" BETWEEN")
	self.appendArguments()

	self.buffer.WriteString(" AND")
	self.appendArguments()
	self.buffer.WriteString(")")

	self.params = append(self.params, column.Type.ToExternal(v1))
	self.params = append(self.params, column.Type.ToExternal(v2))
	return nil
}

func (self *whereBuilder) equalClass(column string, table *types.TableDefinition) {
	if !table.IsInheritanced() {
		return
	}

	if nil == table.Super {
		return
	}
	self.equalsTable(column, table)
}

func (self *whereBuilder) equalsTable(column string, table *types.TableDefinition) {

	cm := table.UnderscoreName
	if !table.HasChildren() {
		self.append(column, "= '"+cm+"'")
		return
	}

	if self.isFirst {
		self.isFirst = false
		if 0 != len(self.prefix) {
			self.buffer.WriteString(self.prefix)
		}
	} else {
		self.buffer.WriteString(" AND ")
	}

	self.buffer.WriteString(column)
	self.buffer.WriteString(" IN ( '")
	self.buffer.WriteString(table.UnderscoreName)
	self.buffer.WriteString("'")

	for _, child := range table.OwnChildren.All() {
		self.buffer.WriteString(", '")
		self.buffer.WriteString(child.UnderscoreName)
		self.buffer.WriteString("'")
	}
	self.buffer.WriteString(")")
}

func (self *whereBuilder) equals_class(column *types.ColumnDefinition, s string) error {
	table := self.tables.FindByUnderscoreName(s)
	if nil == table {
		return errors.New("table '" + s + "' is undefined for column '" + column.Name + "'.")
	}

	self.equalsTable(column.Name, table)
	return nil
}

func (self *whereBuilder) in_class(column *types.ColumnDefinition, s string) error {
	return errors.New("in_class is not implemented.")
}

func (self *whereBuilder) equals_parentClass(column *types.ColumnDefinition, s string) error {
	table := self.tables.FindByUnderscoreName(s)
	if nil == table {
		return errors.New("table '" + s + "' is undefined for column '" + column.Name + "'.")
	}

	self.equalsTable(column.Name, table)
	return nil
}

func (self *whereBuilder) in_parentClass(column *types.ColumnDefinition, s string) error {
	return errors.New("in_parentClass is not implemented.")
}

func split(exp string) (string, string) {
	if '[' != exp[0] {
		return "eq", exp
	}

	idx := strings.IndexRune(exp[1:], ']')
	if -1 == idx {
		return "eq", exp
	}
	return exp[1 : idx+1], exp[idx+2:]
}

func (self *whereBuilder) build(params map[string]string) error {
	if nil == params || 0 == len(params) {
		return nil
	}

	attributes := self.table.Attributes
	for nm, exp := range params {
		if 0 == len(nm) {
			continue
		}

		if '@' != nm[0] {
			continue
		}

		name := nm[1:]
		if ']' == name[len(name)-1] {
			idx := strings.LastIndex(name, "[")
			if -1 != idx {
				name = name[:idx]
			}
		}

		column, _ := attributes[name]
		if nil == column {
			return errors.New("column '" + name + "' is not exists in the " + self.table.Name + ".")
		}

		op, v := split(exp)

		operators := self.operators_for_field[column.Name]
		if nil == operators {
			operators = self.operators
		}

		f, _ := operators[op]
		if nil == f {
			return errors.New("'" + op + "' is unsupported operator for the column '" + name + "'.")
		}
		e := f(self, column, v)
		if nil != e {
			return e
		}
	}
	return nil
}

func (self *whereBuilder) buildSQL(params map[string]string) error {
	e := self.build(params)
	if nil != e {
		return e
	}

	if groupBy, ok := params["group_by"]; ok {
		if 0 == len(groupBy) {
			return errors.New("groupBy is empty.")
		}

		self.buffer.WriteString(" GROUP BY ")
		self.buffer.WriteString(groupBy)
	} else if _, ok := params["group_by[0]"]; ok {
		return errors.New("groupBy have multiple occurrences in the query.")
	}

	if having, ok := params["having"]; ok {
		if 0 == len(having) {
			return errors.New("having is empty.")
		}

		self.buffer.WriteString(" HAVING ")
		self.buffer.WriteString(having)
	} else if _, ok := params["having[0]"]; ok {
		return errors.New("having have multiple occurrences in the query.")
	}

	if order, ok := params["order"]; ok {
		if 0 == len(order) {
			return errors.New("order is empty.")
		}

		self.buffer.WriteString(" ORDER BY ")
		self.buffer.WriteString(order)
	} else if _, ok := params["order[0]"]; ok {
		return errors.New("order have multiple occurrences in the query.")
	}

	if limit, ok := params["limit"]; ok {
		i, e := strconv.ParseInt(limit, 10, 64)
		if nil != e {
			return fmt.Errorf("limit is not a number, actual value is '" + limit + "'")
		}
		if i <= 0 {
			return fmt.Errorf("limit must is geater zero, actual value is '" + limit + "'")
		}

		if nil == self.limit_and_offset {
			self.limit_and_offset = (*whereBuilder).limit_and_offset_generic
		}

		if offset, ok := params["offset"]; ok {
			i, e = strconv.ParseInt(offset, 10, 64)
			if nil != e {
				return fmt.Errorf("offset is not a number, actual value is '" + offset + "'")
			}

			if i < 0 {
				return fmt.Errorf("offset must is geater(or equals) zero, actual value is '" + offset + "'")
			}

			self.limit_and_offset(self, limit, offset)
		} else if _, ok := params["offset[0]"]; ok {
			return errors.New("offset have multiple occurrences in the query.")
		} else {
			self.limit_and_offset(self, limit, "")
		}
	} else if _, ok := params["limit[0]"]; ok {
		return errors.New("limit have multiple occurrences in the query.")
	}
	return nil
}

var index_string = init_index_string()

func init_index_string() []string {
	res := make([]string, 100)
	for i := 0; i < len(res); i++ {
		res[i] = "[" + strconv.FormatInt(int64(i), 10) + "]"
	}
	return res
}

func convertQueryParams(params map[string][]string) map[string]string {
	if nil == params {
		return nil
	}
	if 0 == len(params) {
		return map[string]string{}
	}
	res := make(map[string]string, 2*len(params))
	for k, vv := range params {
		switch len(vv) {
		case 0:
			break
		case 1:
			res[k] = vv[0]
		case 2:
			res[k+"[0]"] = vv[0]
			res[k+"[1]"] = vv[1]
		case 3:
			res[k+"[0]"] = vv[0]
			res[k+"[1]"] = vv[1]
			res[k+"[2]"] = vv[2]
		case 4:
			res[k+"[0]"] = vv[0]
			res[k+"[1]"] = vv[1]
			res[k+"[2]"] = vv[2]
			res[k+"[3]"] = vv[3]
		case 5:
			res[k+"[0]"] = vv[0]
			res[k+"[1]"] = vv[1]
			res[k+"[2]"] = vv[2]
			res[k+"[3]"] = vv[3]
			res[k+"[4]"] = vv[4]
		default:
			for idx, v := range vv {
				if idx < len(index_string) {
					res[k+index_string[idx]] = v
				} else {
					res[k+"["+strconv.FormatInt(int64(idx), 10)+"]"] = v
				}
			}
		}
	}
	return res
}

func (self *whereBuilder) buildWithQueryParams(params map[string][]string) error {
	if nil == params || 0 == len(params) {
		return nil
	}

	attributes := self.table.Attributes
	for nm, exps := range params {
		if '@' != nm[0] {
			continue
		}
		for idx, exp := range exps {
			column, _ := attributes[nm[1:]]
			if nil == column {
				return errors.New("column '" + nm[1:] + "[" + strconv.FormatInt(int64(idx), 10) + "]' is not exists in the " + self.table.Name + ".")
			}

			op, v := split(exp)

			operators := self.operators_for_field[column.Name]
			if nil == operators {
				operators = self.operators
			}

			f, _ := operators[op]
			if nil == f {
				return errors.New("'" + op + "[" + strconv.FormatInt(int64(idx), 10) + "]' is unsupported operator for the column '" + nm[1:] + "'.")
			}
			e := f(self, column, v)
			if nil != e {
				return e
			}
		}
	}
	return nil
}

func (self *whereBuilder) buildSQLWithQueryParams(params map[string][]string) error {
	e := self.buildWithQueryParams(params)
	if nil != e {
		return e
	}

	if groupBy, ok := params["group_by"]; ok {
		if 0 == len(groupBy) {
			return errors.New("groupBy is empty.")
		}
		if 1 != len(groupBy) {
			return errors.New("groupBy have multiple occurrences in the query.")
		}

		self.buffer.WriteString(" GROUP BY ")
		self.buffer.WriteString(groupBy[0])
	}

	if having, ok := params["having"]; ok {
		if 0 == len(having) {
			return errors.New("having is empty.")
		}
		if 1 != len(having) {
			return errors.New("having have multiple occurrences in the query.")
		}

		self.buffer.WriteString(" HAVING ")
		self.buffer.WriteString(having[0])
	}

	if order, ok := params["order"]; ok {
		if 0 == len(order) {
			return errors.New("order is empty.")
		}
		if 1 != len(order) {
			return errors.New("order have multiple occurrences in the query.")
		}

		self.buffer.WriteString(" ORDER BY ")
		self.buffer.WriteString(order[0])
	}

	if limit, ok := params["limit"]; ok {
		if 0 == len(limit) {
			return errors.New("limit is empty.")
		}
		if 1 != len(limit) {
			return errors.New("limit have multiple occurrences in the query.")
		}
		i, e := strconv.ParseInt(limit[0], 10, 64)
		if nil != e {
			return fmt.Errorf("limit is not a number, actual value is '" + limit[0] + "'")
		}
		if i <= 0 {
			return fmt.Errorf("limit must is geater zero, actual value is '" + limit[0] + "'")
		}

		if nil == self.limit_and_offset {
			self.limit_and_offset = (*whereBuilder).limit_and_offset_generic
		}

		if offset, ok := params["offset"]; ok {
			if 0 == len(offset) {
				return errors.New("offset is empty.")
			}
			if 1 != len(offset) {
				return errors.New("offset have multiple occurrences in the query.")
			}

			i, e = strconv.ParseInt(offset[0], 10, 64)
			if nil != e {
				return fmt.Errorf("offset is not a number, actual value is '" + offset[0] + "'")
			}

			if i < 0 {
				return fmt.Errorf("offset must is geater(or equals) zero, actual value is '" + offset[0] + "'")
			}

			self.limit_and_offset(self, limit[0], offset[0])
		} else {
			self.limit_and_offset(self, limit[0], "")
		}
	}
	return nil
}

func newWhere(drv string, table *types.TableDefinition, idx int) *whereBuilder {
	builder := &whereBuilder{tables: nil,
		table:               table,
		idx:                 idx,
		isFirst:             true,
		prefix:              " WHERE ",
		buffer:              bytes.NewBuffer(make([]byte, 0, 100)),
		operators:           default_operators,
		operators_for_field: nil,
		add_argument:        (*whereBuilder).appendNumericArguments}

	if IsNumericParams(drv) {
		builder.add_argument = (*whereBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*whereBuilder).appendSimpleArguments
	}

	switch GetDBType(drv) {
	case POSTGRESQL:
		builder.limit_and_offset = (*whereBuilder).limit_and_offset_postgres
	default:
		builder.limit_and_offset = (*whereBuilder).limit_and_offset_generic
	}

	// POSTGRESQL  = 1
	// MSSQL       = 2
	// ORACLE      = 3
	// SQLITE      = 4
	// MYSQL       = 5
	return builder
}

func BuildWhereWithQueryParams(drv string, table *types.TableDefinition, idx int, params map[string][]string) (string, []interface{}, error) {
	if nil == params || 0 == len(params) {
		return "", nil, nil
	}

	builder := newWhere(drv, table, idx)
	e := builder.buildSQLWithQueryParams(params)
	if nil != e {
		return "", nil, e
	}

	return builder.buffer.String(), builder.params, nil
}

func BuildWhere(drv string, table *types.TableDefinition, idx int, params map[string]string) (string, []interface{}, error) {
	if nil == params || 0 == len(params) {
		return "", nil, nil
	}

	builder := newWhere(drv, table, idx)
	e := builder.buildSQL(params)
	if nil != e {
		return "", nil, e
	}

	return builder.buffer.String(), builder.params, nil
}

type updateBuilder struct {
	drv    *simple_driver
	table  *types.TableDefinition
	idx    int
	buffer *bytes.Buffer
	params []interface{}
}

func (self *updateBuilder) buildUpdate(updated_attributes map[string]interface{}) error {
	self.buffer.WriteString("UPDATE ")

	if self.drv.dbType == POSTGRESQL && self.drv.hasOnly {
		self.buffer.WriteString(" ONLY ")
	}
	self.buffer.WriteString(self.table.CollectionName)
	self.buffer.WriteString(" SET ")
	isFirst := true
	var e error

	for _, attribute := range self.table.Attributes {
		//////////////////////////////////////////
		// TODO: refactor it?
		if attribute.IsSerial() {
			continue
		}
		var value interface{} = nil
		switch attribute.Name {
		case "created_at":
			continue
		case "updated_at":
			value = time.Now()
		case "type":
			if _, ok := updated_attributes[attribute.Name]; ok {
				return errors.New("column 'type' is readonly.")
			}
			continue
		default:
			v := updated_attributes[attribute.Name]
			if nil == v {
				continue
			}
			if attribute.IsReadOnly {
				return fmt.Errorf("column '%v' is readonly.", attribute.Name)
			}
			value, e = attribute.Type.ToInternal(v)
			if nil != e {
				return fmt.Errorf("column '%v' is not a '%v', actual value is '%v'",
					attribute.Name, attribute.Type.Name(), v)
			}
			value = attribute.Type.ToExternal(value)
		}

		//////////////////////////////////////////

		if isFirst {
			isFirst = false
		} else {
			self.buffer.WriteString(", ")
		}

		self.buffer.WriteString(attribute.Name)
		if self.drv.isNumericParams {
			self.buffer.WriteString("= $")
			self.buffer.WriteString(strconv.FormatInt(int64(self.idx), 10))
		} else {
			self.buffer.WriteString(" = ?")
		}

		self.params = append(self.params, value)

		self.idx++
	}
	return nil
}

func (self *updateBuilder) buildWhereWithString(queryString string, params []interface{}) {
	if 0 == len(queryString) {
		return
	}

	self.buffer.WriteString(" WHERE ")
	if self.drv.isNumericParams {
		self.buffer, self.idx = replaceQuestion(self.buffer, queryString, self.idx)
	} else {
		self.buffer.WriteString(queryString)
	}
	if nil != params && 0 != len(params) {
		self.params = append(self.params, params...)
	}
}

func (self *updateBuilder) buildWhereById(id interface{}) {
	self.buffer.WriteString(" WHERE ")
	self.buffer.WriteString(self.table.Id.Name)
	if self.drv.isNumericParams {
		self.buffer.WriteString(" = $")
		self.buffer.WriteString(strconv.FormatInt(int64(self.idx), 10))
		self.idx++
	} else {
		self.buffer.WriteString(" = ?")
	}

	self.params = append(self.params, self.table.Id.Type.ToExternal(id))
}

func (self *updateBuilder) buildWhere(params map[string]string, isSimpleTableInheritance bool) error {
	builder := self.drv.newWhere(self.idx, self.table, self.buffer)
	builder.params = self.params
	if isSimpleTableInheritance {
		builder.equalClass("type", self.table)
	}

	e := builder.build(params)
	self.params = builder.params
	self.idx = builder.idx
	return e
}
