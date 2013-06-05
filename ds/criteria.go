package ds

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

type op_func func(self *queryBuilder, column *types.ColumnDefinition, v string) error

var (
	default_operators = make(map[string]op_func)
)

func init() {
	default_operators["exists"] = (*queryBuilder).exists
	default_operators["in"] = (*queryBuilder).in
	default_operators["nin"] = (*queryBuilder).nin
	default_operators["gt"] = (*queryBuilder).gt
	default_operators["gte"] = (*queryBuilder).gte
	default_operators["eq"] = (*queryBuilder).eq
	default_operators["ne"] = (*queryBuilder).ne
	default_operators["lt"] = (*queryBuilder).lt
	default_operators["lte"] = (*queryBuilder).lte
	default_operators["between"] = (*queryBuilder).between
	default_operators["is"] = (*queryBuilder).is
	default_operators["like"] = (*queryBuilder).like
}

type queryBuilder struct {
	table   *types.TableDefinition
	idx     int
	isFirst bool
	prefix  string

	buffer *bytes.Buffer
	params []interface{}

	operators    map[string]op_func
	add_argument func(self *queryBuilder)
}

func (self *queryBuilder) appendArguments() {
	if nil != self.add_argument {
		self.add_argument(self)
	} else {
		self.appendSimpleArguments()
	}
}

func (self *queryBuilder) appendNumericArguments() {
	self.buffer.WriteString(" $")
	self.buffer.WriteString(fmt.Sprint(self.idx))
	self.idx++
}

func (self *queryBuilder) appendSimpleArguments() {
	self.buffer.WriteString(" ?")
	self.idx++
}

func (self *queryBuilder) append(ss ...string) error {
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

func (self *queryBuilder) add(column *types.ColumnDefinition, op, s string) error {
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

func (self *queryBuilder) exists(column *types.ColumnDefinition, s string) error {
	return errors.New("not implemented")
}

func (self *queryBuilder) in(column *types.ColumnDefinition, s string) error {
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

func (self *queryBuilder) nin(column *types.ColumnDefinition, s string) error {
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

func (self *queryBuilder) gt(column *types.ColumnDefinition, s string) error {
	return self.add(column, ">", s)
}

func (self *queryBuilder) gte(column *types.ColumnDefinition, s string) error {
	return self.add(column, ">=", s)
}

func (self *queryBuilder) eq(column *types.ColumnDefinition, s string) error {
	return self.add(column, "=", s)
}

func (self *queryBuilder) ne(column *types.ColumnDefinition, s string) error {
	return self.add(column, "!=", s)
}

func (self *queryBuilder) lt(column *types.ColumnDefinition, s string) error {
	return self.add(column, "<", s)
}

func (self *queryBuilder) lte(column *types.ColumnDefinition, s string) error {
	return self.add(column, "<=", s)
}

func (self *queryBuilder) is(column *types.ColumnDefinition, s string) error {
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

func (self *queryBuilder) like(column *types.ColumnDefinition, s string) error {
	if "string" == column.Type.Name() {
		return fmt.Errorf("'like' is not supported for the column '%v', it must is a string type", column.Name)
	}

	return self.append(column.Name, " LIKE '"+s+"'")
}

func (self *queryBuilder) between(column *types.ColumnDefinition, s string) error {
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

func (self *queryBuilder) equalClass(column string, table *types.TableDefinition) {
	if !table.IsInheritanced() {
		return
	}

	if nil == table.Super {
		return
	}

	cm := table.UnderscoreName

	if nil == table.Children || 0 == len(table.Children) {
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

	isFirst := true

	self.buffer.WriteString(column)
	self.buffer.WriteString(" IN (")
	for _, child := range table.Children {
		if isFirst {
			isFirst = false
			self.buffer.WriteString(" '")
		} else {
			self.buffer.WriteString(", '")
		}

		self.buffer.WriteString(", '")
		self.buffer.WriteString(child.UnderscoreName)
		self.buffer.WriteString("'")
	}
	self.buffer.WriteString(")")
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

func (self *queryBuilder) build(params map[string]string) error {
	if nil == params || 0 == len(params) {
		return nil
	}

	attributes := self.table.Attributes
	for nm, exp := range params {
		if '@' != nm[0] {
			continue
		}
		column, _ := attributes[nm[1:]]
		if nil == column {
			return errors.New("column '" + nm[1:] + "' is not exists in the " + self.table.Name + ".")
		}

		op, v := split(exp)
		f, _ := self.operators[op]
		if nil == f {
			return errors.New("'" + op + "' is unsupported operator for the column '" + nm[1:] + "'.")
		}

		e := f(self, column, v)
		if nil != e {
			return e
		}
	}
	return nil
}

type updateBuilder struct {
	table           *types.TableDefinition
	idx             int
	isNumericParams bool
	buffer          *bytes.Buffer
	params          []interface{}
}

func (self *updateBuilder) buildUpdate(updated_attributes map[string]interface{}) error {
	self.buffer.WriteString("UPDATE ")
	self.buffer.WriteString(self.table.CollectionName)
	self.buffer.WriteString(" SET ")
	isFirst := true
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
		default:
			v := updated_attributes[attribute.Name]
			if nil == v {
				continue
			}
			if attribute.IsReadOnly {
				return fmt.Errorf("column '%v' is readonly.", attribute.Name)
			}
			value = attribute.Type.ToExternal(v)
		}

		//////////////////////////////////////////

		if isFirst {
			isFirst = false
		} else {
			self.buffer.WriteString(", ")
		}

		self.buffer.WriteString(attribute.Name)
		if self.isNumericParams {
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
	if self.isNumericParams {
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
	if self.isNumericParams {
		self.buffer.WriteString(" = $")
		self.buffer.WriteString(strconv.FormatInt(int64(self.idx), 10))
		self.idx++
	} else {
		self.buffer.WriteString(" = ?")
	}

	self.params = append(self.params, self.table.Id.Type.ToExternal(id))
}

func (self *updateBuilder) buildWhere(params map[string]string) error {
	builder := &queryBuilder{table: self.table,
		idx:       self.idx,
		isFirst:   true,
		prefix:    " WHERE",
		buffer:    self.buffer,
		params:    self.params,
		operators: default_operators}

	if self.isNumericParams {
		builder.add_argument = (*queryBuilder).appendNumericArguments
	} else {
		builder.add_argument = (*queryBuilder).appendSimpleArguments
	}

	e := builder.build(params)
	self.params = builder.params
	self.idx = builder.idx
	return e
}
