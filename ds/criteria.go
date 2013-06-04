package ds

import (
	"bytes"
	"commons/types"
	"errors"
	"fmt"
	"strings"
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

type op_func func(self *statementBuilder, column *types.ColumnDefinition, v string) error

var (
	default_operators = make(map[string]op_func)
)

func init() {
	default_operators["exists"] = (*statementBuilder).exists
	default_operators["in"] = (*statementBuilder).in
	default_operators["nin"] = (*statementBuilder).nin
	default_operators["gt"] = (*statementBuilder).gt
	default_operators["gte"] = (*statementBuilder).gte
	default_operators["eq"] = (*statementBuilder).eq
	default_operators["ne"] = (*statementBuilder).ne
	default_operators["lt"] = (*statementBuilder).lt
	default_operators["lte"] = (*statementBuilder).lte
	default_operators["between"] = (*statementBuilder).between
	default_operators["is"] = (*statementBuilder).is
	default_operators["like"] = (*statementBuilder).like
}

type statementBuilder struct {
	table   *types.TableDefinition
	idx     int
	isFirst bool

	buffer *bytes.Buffer
	params []interface{}

	operators    map[string]op_func
	add_argument func(self *statementBuilder)
}

func (self *statementBuilder) appendArguments() {
	if nil != self.add_argument {
		self.add_argument(self)
	} else {
		self.appendSimpleArguments()
	}
}

func (self *statementBuilder) appendNumericArguments() {
	self.buffer.WriteString(" $")
	self.buffer.WriteString(fmt.Sprint(self.idx))
	self.idx++
}

func (self *statementBuilder) appendSimpleArguments() {
	self.buffer.WriteString(" ?")
	self.idx++
}

func (self *statementBuilder) append(ss ...string) error {
	if 0 == len(ss) {
		return nil
	}
	if self.isFirst {
		self.isFirst = false
	} else {
		self.buffer.WriteString(" AND ")
	}
	for _, s := range ss {
		self.buffer.WriteString(" ")
		self.buffer.WriteString(s)
	}
	return nil
}

func (self *statementBuilder) add(column *types.ColumnDefinition, op, s string) error {
	v, e := column.Type.Parse(s)
	if nil != e {
		return fmt.Errorf("column '%v' convert '%v' to '%v' failed, %v",
			column.Name, s, column.Type.Name(), e)
	}

	if self.isFirst {
		self.isFirst = false
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

func (self *statementBuilder) exists(column *types.ColumnDefinition, s string) error {
	return errors.New("not implemented")
}

func (self *statementBuilder) in(column *types.ColumnDefinition, s string) error {
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

func (self *statementBuilder) nin(column *types.ColumnDefinition, s string) error {
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

func (self *statementBuilder) gt(column *types.ColumnDefinition, s string) error {
	return self.add(column, ">", s)
}

func (self *statementBuilder) gte(column *types.ColumnDefinition, s string) error {
	return self.add(column, ">=", s)
}

func (self *statementBuilder) eq(column *types.ColumnDefinition, s string) error {
	return self.add(column, "=", s)
}

func (self *statementBuilder) ne(column *types.ColumnDefinition, s string) error {
	return self.add(column, "!=", s)
}

func (self *statementBuilder) lt(column *types.ColumnDefinition, s string) error {
	return self.add(column, "<", s)
}

func (self *statementBuilder) lte(column *types.ColumnDefinition, s string) error {
	return self.add(column, "<=", s)
}

func (self *statementBuilder) is(column *types.ColumnDefinition, s string) error {
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

func (self *statementBuilder) like(column *types.ColumnDefinition, s string) error {
	if "string" == column.Type.Name() {
		return fmt.Errorf("'like' is not supported for the column '%v', it must is a string type", column.Name)
	}

	return self.append(column.Name, " LIKE '"+s+"'")
}

func (self *statementBuilder) between(column *types.ColumnDefinition, s string) error {
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

func (self *statementBuilder) equalClass(column string, table *types.TableDefinition) {
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

func (self *statementBuilder) build(params map[string]string) error {
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
