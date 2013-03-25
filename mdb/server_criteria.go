package mdb

import (
	"commons"
	"commons/stringutils"
	"errors"
	"labix.org/v2/mgo/bson"
	"strings"
)

// func collectOwnProperties(cls *ClassDefinition, properties map[string]*PropertyDefinition) {

// 	for k, p := range cls.OwnProperties {
// 		properties[k] = p
// 	}

// 	if nil != cls.Children || 0 == len(cls.Children) {
// 		return
// 	}

// 	for _, child := range cls.Children {
// 		collectOwnProperties(child, properties)
// 	}
// }

// func collectProperties(cls *ClassDefinition) map[string]*PropertyDefinition {
// 	if nil != cls.Children || 0 == len(cls.Children) {
// 		return cls.Properties
// 	}

// 	properties := make(map[string]*PropertyDefinition, len(cls.Properties))

// 	for k, p := range cls.Properties {
// 		properties[k] = p
// 	}

// 	for _, child := range cls.Children {
// 		collectOwnProperties(child, properties)
// 	}
// 	return properties
// }

func parseObjectIdHex(s string) (id bson.ObjectId, err error) {
	defer func() {
		if e := recover(); nil != e {
			err = commons.NewError(e)
		}
	}()

	v := bson.ObjectIdHex(s)
	return v, nil
}

func appendIdCriteria(q bson.M, exp string) error {
	var err error
	var cr interface{}
	if '[' != exp[0] {
		cr, err = parseObjectIdHex(exp)
	} else if strings.HasPrefix(exp, "[eq]") {
		cr, err = parseObjectIdHex(exp[4:])
	} else {
		return errors.New("invalid operator for _id - " + exp)
	}
	if nil != err {
		return errors.New("_id is a invalid ObjectId")
	}
	q["_id"] = cr
	return nil
}

// func findPropertyDefinitionIfIsArrayName(cls *ClassDefinition, nm string) (*PropertyDefinition, error) {
// 	pos := strings.LastIndex(nm, ".")
// 	if -1 != pos {
// 		pr, _ = properties[nm[0:pos]]
// 	}

// 	if nil == pr {

// 		if is_all {
// 			return nil, errors.New("'" + nm + "' is not a property.")
// 		}
// 		properties = collectProperties(cls)
// 		is_all = true
// 		pr, _ = properties[nm]

// 		if nil == pr {
// 			if -1 != pos {
// 				pr, _ = properties[nm[0:pos]]
// 			}

// 			if nil == pr {
// 				return nil, errors.New("'" + nm + "' is not a property.")
// 			}
// 		}
// 	}
// }

func buildInheritanceQuery(cls *ClassDefinition) bson.M {
	if !cls.IsInheritance() {
		return nil
	}
	if nil == cls.Super {
		return nil
	}
	cm := stringutils.Underscore(cls.Name)
	if nil == cls.Children || 0 == len(cls.Children) {
		return bson.M{"type": cm}
	}
	ar := make([]interface{}, 0, len(cls.Children))
	ar = append(ar, cm)
	for _, child := range cls.Children {
		ar = append(ar, stringutils.Underscore(child.Name))
	}
	return bson.M{"type": bson.M{"$in": ar}}
}

func buildClassQuery(cls *ClassDefinition) interface{} {
	cm := stringutils.Underscore(cls.Name)
	if !cls.IsInheritance() {
		return cm
	}
	if nil == cls.Children || 0 == len(cls.Children) {
		return cm
	}

	ar := make([]interface{}, 0, len(cls.Children))
	ar = append(ar, cm)
	for _, child := range cls.Children {
		ar = append(ar, stringutils.Underscore(child.Name))
	}
	return bson.M{"$in": ar}
}

func (self *mdb_server) buildClassQueryFromClassName(t string) (interface{}, error) {
	cls := self.definitions.FindByUnderscoreName(t)
	if nil == cls {
		return nil, errors.New("class '" + t + "' is not found")
	}
	return buildClassQuery(cls), nil
}

func (self *mdb_server) buildQueryStatement(cls *ClassDefinition, params map[string]string) (bson.M, error) {
	q := buildInheritanceQuery(cls)
	if nil == params || 0 == len(params) {
		return q, nil
	}

	//is_all := nil != cls.Children || 0 == len(cls.Children)
	properties := cls.Properties
	if nil == q {
		q = bson.M{}
	}

	for nm, exp := range params {
		if '@' != nm[0] {
			continue
		}
		nm = nm[1:]

		if "parent_type" == nm {
			s, e := self.buildClassQueryFromClassName(exp)
			if nil != e {
				return nil, e
			}
			if nil != s {
				q[nm] = s
			}
			continue
		}

		pr, _ := properties[nm]
		if nil == pr {
			if "_id" == nm {
				e := appendIdCriteria(q, exp)
				if nil != e {
					return nil, e
				}
				continue
			}

			pos := strings.LastIndex(nm, ".")
			if -1 == pos {
				return nil, errors.New("'" + nm + "' is not a property in " + cls.String() + ".")
			}

			pr, _ = properties[nm[0:pos]]
			if nil == pr {
				return nil, errors.New("'" + nm + "' is not a property in " + cls.String() + ".")
			}
		}

		var ss []string
		if '[' == exp[0] {
			ss = strings.SplitN(exp[1:], "]", 2)
		} else {
			ss = nil
		}

		if nil == ss || 2 != len(ss) {
			v, err := pr.Type.Convert(exp)
			if nil != err {
				return nil, errors.New("'" + nm + "' convert to " +
					pr.Type.Name() + ", failed, " + err.Error())
			}
			q[nm] = v
			continue
		}

		f, _ := operators[ss[0]]
		if nil == f {
			return nil, errors.New(" '" + ss[0] + "' is unsupported operator for '" +
				nm + "'.")
		}
		value, err := f(pr, ss[1])
		if nil != err {
			return nil, errors.New("'" + nm + "' convert to " +
				pr.Type.Name() + ", failed, " + err.Error())
		}
		q[nm] = value
	}
	if 0 == len(q) {
		return nil, nil
	}
	return q, nil
}
