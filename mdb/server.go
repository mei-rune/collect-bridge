package mdb

import (
	"commons"
	"commons/as"
	"commons/stringutils"
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
	"time"
)

var (
	assocationOps = make([]*assocationOp, 5)
	operators     = make(map[string]func(pr *PropertyDefinition, s string) (interface{}, error))
)

type assocationOp struct {
	deleteOp    func(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error
	deleteAllOp func(s *mdb_server, assoc Assocation, cls *ClassDefinition) error
	//createOp func(s *mdb_server, assoc *Assocation, id interface{}) error
}

func init() {
	assocationOps[BELONGS_TO] = &assocationOp{}
	assocationOps[HAS_ONE] = &assocationOp{deleteOp: deleteHasMany, deleteAllOp: deleteAllHasMany}
	assocationOps[HAS_MANG] = &assocationOp{deleteOp: deleteHasMany, deleteAllOp: deleteAllHasMany}
	assocationOps[HAS_AND_BELONGS_TO_MANY] = &assocationOp{deleteOp: deleteMany2Many, deleteAllOp: deleteAllMany2Many}

	operators["exists"] = op_exist
	operators["in"] = op_in
	operators["nin"] = op_nin
	operators["gt"] = op_gt
	operators["gte"] = op_gte
	operators["eq"] = op_eq
	operators["ne"] = op_ne
	operators["lt"] = op_lt
	operators["lte"] = op_lte

	errors.New("")
}

func op_exist(pr *PropertyDefinition, s string) (interface{}, error) {
	switch s {
	case "true":
		return bson.M{"$exists": true}, nil
	case "false":
		return bson.M{"$exists": false}, nil
	}
	return nil, errors.New("'exist' of '" + pr.Name + "' require one bool operand - " + s)
}
func op_in(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_nin(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_gt(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$gt": v}, nil
}
func op_gte(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$gte": v}, nil
}
func op_eq(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return v, nil
}
func op_ne(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$ne": v}, nil
}
func op_lt(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$lt": v}, nil
}
func op_lte(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$lte": v}, nil
}

type mdb_server struct {
	session     *mgo.Database
	restrict    bool
	definitions *ClassDefinitions
}

func checkValue(pr *PropertyDefinition, attributes map[string]interface{}, value interface{}, errs *[]error) (interface{}, bool) {

	new_value, err := pr.Type.Convert(value)
	if nil != err {
		*errs = append(*errs, errors.New("'"+pr.Name+"' convert to internal value failed, "+err.Error()))
		return nil, true
	}

	if nil != pr.Restrictions && 0 != len(pr.Restrictions) {
		is_failed := false
		for _, r := range pr.Restrictions {
			if ok, err := r.Validate(new_value, attributes); !ok {
				*errs = append(*errs, errors.New("'"+pr.Name+"' is validate failed, "+err.Error()))
				is_failed = true
			}
		}

		if is_failed {
			return nil, true
		}
	}
	return new_value, false
}

func doMagic(k string, attributes, new_attributes map[string]interface{},
	is_update bool, errs *[]error) bool {
	if k == "updated_at" {
		_, ok := attributes[k]
		if ok {
			*errs = append(*errs, errors.New("'"+k+"' is magic property"))
			return true
		}

		new_attributes[k] = time.Now()
		return true
	}

	if k == "created_at" {
		_, ok := attributes[k]
		if ok {
			*errs = append(*errs, errors.New("'"+k+"' is magic property"))
			return true
		}

		if !is_update {
			new_attributes[k] = time.Now()
		}
		return true
	}
	return false
}
func doInheritance(cls *ClassDefinition, attributes, new_attributes map[string]interface{},
	is_update bool, errs *[]error) {
	t, ok := attributes["type"]
	if ok {
		if !cls.IsInheritance() {
			*errs = append(*errs, errors.New("it is not inheritance and cannot contains 'type'"))
			return
		}
		cm := stringutils.Underscore(cls.Name)
		if cm != t {
			*errs = append(*errs, fmt.Errorf("'%t' of 'type' is not equal class '%s'", t, cm))
			return
		}
		new_attributes["type"] = t
	} else {
		if cls.IsInheritance() {
			new_attributes["type"] = stringutils.Underscore(cls.Name)
		}
	}
}

func (self *mdb_server) preWrite(cls *ClassDefinition, uattributes map[string]interface{},
	is_update bool) (map[string]interface{}, error) {
	attributes := uattributes
	if self.restrict {
		attributes = make(map[string]interface{}, len(attributes))
		for k, v := range uattributes {
			attributes[k] = v
		}
	}

	new_attributes := make(map[string]interface{}, len(attributes))
	errs := make([]error, 0, 10)

	doInheritance(cls, attributes, new_attributes, is_update, &errs)

	for k, pr := range cls.Properties {

		if doMagic(k, attributes, new_attributes, is_update, &errs) {
			continue
		}

		value, ok := attributes[k]
		if !ok {
			if COLLECTION_UNKNOWN != pr.Collection {
				continue
			}

			if is_update {
				continue
			}

			if pr.IsRequired {
				errs = append(errs, errors.New("'"+k+"' is required"))
				continue
			}
			if nil == pr.DefaultValue {
				continue
			}

			new_attributes[k] = pr.DefaultValue
			continue
		}

		if pr.IsReadOnly {
			errs = append(errs, errors.New("'"+k+"' is readonly"))
			continue
		}

		if self.restrict {
			delete(attributes, k)
		}

		if COLLECTION_UNKNOWN == pr.Collection {
			new_value, is_failed := checkValue(pr, uattributes, value, &errs)

			if is_failed {
				continue
			}

			new_attributes[k] = new_value
			continue
		}

		array, ok := value.([]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("'"+k+"' must is a collection, actual is %v", value))
			continue
		}
		is_failed := false
		new_array := make([]interface{}, 0, len(array))
		for _, v := range array {
			nv, failed := checkValue(pr, uattributes, v, &errs)
			if failed {
				is_failed = true
			} else {
				new_array = append(new_array, nv)
			}
		}

		if is_failed {
			continue
		}

		new_attributes[k] = new_array
	}

	if 0 != len(errs) {
		return nil, commons.NewMutiErrors("validate failed", errs)
	}

	if self.restrict && 0 != len(attributes) {
		for k, _ := range attributes {
			if '$' == k[0] { // it is child of the model
				continue
			}
			errs = append(errs, errors.New("'"+k+"' is useless"))
		}
		return nil, commons.NewMutiErrors("validate failed", errs)
	}
	return new_attributes, nil
}

func (self *mdb_server) postRead(raw_cls *ClassDefinition, attributes map[string]interface{}) (map[string]interface{}, error) {

	new_attributes := make(map[string]interface{}, len(attributes))

	cls := raw_cls
	if raw_cls.IsInheritance() {
		t, ok := attributes["type"]
		if !ok {
			return nil, fmt.Errorf("result is not contains 'type' - %v", attributes)
		}

		nm, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("'type' is not a string type, it is  '%T:%s' is not found - %v", t, t, attributes)
		}

		cls = self.definitions.FindByUnderscoreName(nm)
		if nil == cls {
			return nil, fmt.Errorf("class '%s' is not found  - %v", nm, attributes)
		}
		if !cls.InheritanceFrom(raw_cls) {
			return nil, fmt.Errorf("class '%s' is not child of class %s  - %v", cls.Name, raw_cls.Name, attributes)
		}
		new_attributes["type"] = nm
	}

	errs := make([]error, 0, 10)
	for k, pr := range cls.Properties {
		value, ok := attributes[k]
		if !ok {
			if nil != pr.DefaultValue {
				new_attributes[k] = pr.DefaultValue
			}
			continue
		}

		if COLLECTION_UNKNOWN == pr.Collection {
			new_value, err := pr.Type.Convert(value)
			if nil != err {
				errs = append(errs, fmt.Errorf("value '%v' of key '%s' convert to internal value failed, %s", value, k, err.Error()))
			} else {
				new_attributes[k] = new_value
			}
			continue
		}
		array, _ := as.AsArray(value)
		if nil == array {
			errs = append(errs, fmt.Errorf("'"+k+"' must is a collection, actual is %v", value))
			continue
		}

		new_array := make([]interface{}, 0, len(array))
		is_failed := false
		for _, v := range array {
			nv, err := pr.Type.Convert(v)
			if nil != err {
				errs = append(errs, fmt.Errorf("value '%v' of key '%s' convert to internal value failed, %s", value, k, err.Error()))
				is_failed = true
			} else {
				new_array = append(new_array, nv)
			}
		}

		if is_failed {
			continue
		}
		new_attributes[k] = new_array
	}

	if t, ok := attributes["type"]; ok {
		new_attributes["type"] = t
	}

	if 0 != len(errs) {
		return nil, commons.NewMutiErrors("validate failed", errs)
	}
	return new_attributes, nil
}

func (self *mdb_server) findClassByAttributes(attributes map[string]interface{}) (*ClassDefinition, error) {
	if nil == attributes {
		return nil, nil
	}
	objectType, ok := attributes["type"]
	if !ok {
		return nil, nil
	}

	t, ok := objectType.(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("type '%v' in body is not a string type", objectType))
	}

	definition := self.definitions.FindByUnderscoreName(t)
	if nil == definition {
		return nil, errors.New("class '" + t + "' is not found")
	}
	return definition, nil
}

func (self *mdb_server) findClass(t string, attributes map[string]interface{}) (*ClassDefinition, error) {
	cls, e := self.findClassByAttributes(attributes)
	if nil != e {
		return nil, e
	}
	if nil != cls {
		return cls, nil
	}
	definition := self.definitions.FindByUnderscoreName(t)
	if nil == definition {
		return nil, errors.New("class '" + t + "' is not found")
	}
	return definition, nil
}

func (self *mdb_server) createChildren(cls *ClassDefinition, id interface{}, attributes map[string]interface{}) []string {
	warnings := []string{}
	for k, v := range attributes {
		if '$' != k[0] {
			continue
		}
		ccls := self.definitions.FindByUnderscoreName(k[1:])
		if nil == ccls {
			warnings = append(warnings, fmt.Sprintf("class '%s' of '%s' is not found", k[1:], k))
			continue
		}

		assoc := cls.GetAssocationByCollectionName(ccls.CollectionName())
		if nil == assoc {
			warnings = append(warnings, fmt.Sprintf("class '%s' is not contains child that name is '%s' at the '%s'", cls.Name, k[1:], k))
			continue
		}
		foreignKey := ""
		is_polymorphic := false
		switch a := assoc.(type) {
		case *HasMany:
			foreignKey = a.ForeignKey
			is_polymorphic = a.Polymorphic
		case *HasOne:
			foreignKey = a.ForeignKey
		default:
			warnings = append(warnings, fmt.Sprintf("class '%s' is not contains child that name is '%s' at the '%s'", cls.Name, k[1:], k))
		}

		commons.Each(v, func(ck interface{}, r interface{}) {
			attrs, ok := r.(map[string]interface{})
			if !ok {
				warnings = append(warnings, fmt.Sprintf("value of '%s.%s' is not map[string]interface{}", k, ck))
				return
			}

			lcls, err := self.findClassByAttributes(attrs)
			if nil != err {
				warnings = append(warnings, fmt.Sprintf("class '%s' of '%s.%s' is not found, %v - %v", k[1:], k, ck, err, attrs))
				return
			}
			if nil == lcls {
				lcls = ccls
			}
			if is_polymorphic {
				attrs["parent_type"] = stringutils.Underscore(cls.Name)
				attrs["parent_id"] = id
			} else {
				attrs[foreignKey] = id
			}
			_, err = self.Create(lcls, attrs)
			if nil != err {
				warnings = append(warnings, fmt.Sprintf("save '%s.%s' failed, %v - %v", k, ck, err, attrs))
			}
		}, func() {
			warnings = append(warnings, fmt.Sprintf("value of '%s' is not []interface{} or map[string]interface{}", k))
		})
	}
	if 0 == len(warnings) {
		return nil
	}
	return warnings
}

func (self *mdb_server) Create(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	new_attributes, errs := self.preWrite(cls, attributes, false)
	if nil != errs {
		return nil, errs
	}

	id, ok := new_attributes["_id"]
	if !ok {
		id = bson.NewObjectId()
		new_attributes["_id"] = id
	}
	err := self.session.C(cls.CollectionName()).Insert(new_attributes)
	if nil != err {
		return nil, err
	}

	return id, nil
}

func (self *mdb_server) FindById(cls *ClassDefinition, id interface{}) (map[string]interface{}, error) {
	var q *mgo.Query
	var result map[string]interface{}

	q = self.session.C(cls.CollectionName()).FindId(id)
	if nil == q {
		return nil, errors.New("return nil result")
	}
	err := q.One(&result)
	if nil != err {
		return nil, err
	}

	return self.postRead(cls, result)
}

func (self *mdb_server) Update(cls *ClassDefinition, id interface{}, updated_attributes map[string]interface{}) error {
	new_attributes, errs := self.preWrite(cls, updated_attributes, true)
	if nil != errs {
		return errs
	}

	return self.session.C(cls.CollectionName()).UpdateId(id, bson.M{"$set": new_attributes})
}

func collectionExists(s *mdb_server, cn string) bool {
	names, e := s.session.CollectionNames()
	if nil != e {
		return false
	}
	for _, nm := range names {
		if nm == cn {
			return true
		}
	}
	return false
}

func deleteHasMany(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	var qc bson.M
	if hasMany.Polymorphic {
		qc = bson.M{"parent_type": cls.CollectionName(), "parent_id": id}
	} else {
		qc = bson.M{hasMany.ForeignKey: id}
	}
	it := s.session.C(hasMany.Target().CollectionName()).Find(qc).Select(bson.M{"_id": 1}).Iter()
	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}

		s.RemoveById(assoc.Target(), o)
	}
	return it.Err()
}
func deleteAllHasMany(s *mdb_server, assoc Assocation, cls *ClassDefinition) error {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	cn := hasMany.Target().CollectionName()
	if hasMany.Polymorphic {
		_, err := s.RemoveBy(hasMany.Target(), map[string]string{"@parent_type": cls.CollectionName()})
		if nil != err {
			return fmt.Errorf("delete from '%s' collection failed, %v", cn, err)
		}
		return nil
	}
	_, err := s.RemoveAll(hasMany.Target(), map[string]string{})
	if nil != err {
		if !collectionExists(s, cn) {
			return nil
		}
		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
	}
	return nil
}

func deleteMany2Many(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error {
	habtm, ok := assoc.(*HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}
	it := s.session.C(habtm.CollectionName).Find(bson.M{habtm.ForeignKey: id}).Select(bson.M{"_id": 1}).Iter()

	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}
		s.RemoveById(assoc.Target(), o)
	}

	return it.Err()
}

func deleteAllMany2Many(s *mdb_server, assoc Assocation, cls *ClassDefinition) error {
	habtm, ok := assoc.(*HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}
	cn := habtm.CollectionName
	err := s.session.C(cn).DropCollection()
	if nil != err {
		if !collectionExists(s, cn) {
			return nil
		}
		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
	}
	return nil
}

func (self *mdb_server) RemoveById(cls *ClassDefinition, id interface{}) (bool, error) {
	err := self.session.C(cls.CollectionName()).RemoveId(id)
	if nil != err {
		return false, err
	}
	_, err = self.removeChildren(cls, id)
	return true, err
}

func (self *mdb_server) removeChildren(cls *ClassDefinition, id interface{}) (bool, error) {
	errs := make([]error, 0)
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteOp {
			continue
		}
		err := op.deleteOp(self, a, cls, id)
		if nil != err {
			errs = append(errs, err)
		}
	}
	if 0 == len(errs) {
		return true, nil
	}
	return false, commons.NewMutiErrors("parameters is error.", errs)
}

func (self *mdb_server) removeAllChildren(cls *ClassDefinition) error {
	errs := make([]error, 0)
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteAllOp {
			continue
		}
		err := op.deleteAllOp(self, a, cls)
		if nil != err {
			errs = append(errs, err)
		}
	}
	if 0 == len(errs) {
		return nil
	}
	return commons.NewMutiErrors("parameters is error.", errs)
}

func (self *mdb_server) RemoveBy(cls *ClassDefinition, params map[string]string) (bool, error) {
	s, err := buildQueryStatement(cls, params)
	if nil != err {
		return false, err
	}
	c := self.session.C(cls.CollectionName())
	q := c.Find(s)
	if nil == q {
		return false, errors.New("return nil result")
	}
	iter := q.Select(bson.M{"_id": 1}).Iter()
	var result map[string]interface{}
	for iter.Next(&result) {
		self.removeChildren(cls, result["_id"])
	}

	err = iter.Err()
	if nil != err {
		return false, err
	}
	_, err = c.RemoveAll(s)
	return (nil == err), err
}

func (self *mdb_server) RemoveAll(cls *ClassDefinition, params map[string]string) (bool, error) {
	err := self.removeAllChildren(cls)
	if nil != err {
		return false, err
	}
	err = self.session.C(cls.CollectionName()).DropCollection()
	if nil != err {
		if !collectionExists(self, cls.CollectionName()) {
			return true, nil
		}

		return false, err
	}
	return true, nil
}

func collectOwnProperties(cls *ClassDefinition, properties map[string]*PropertyDefinition) {

	for k, p := range cls.OwnProperties {
		properties[k] = p
	}

	if nil != cls.Children || 0 == len(cls.Children) {
		return
	}

	for _, child := range cls.Children {
		collectOwnProperties(child, properties)
	}
}

func collectProperties(cls *ClassDefinition) map[string]*PropertyDefinition {
	if nil != cls.Children || 0 == len(cls.Children) {
		return cls.Properties
	}

	properties := make(map[string]*PropertyDefinition, len(cls.Properties))

	for k, p := range cls.Properties {
		properties[k] = p
	}

	for _, child := range cls.Children {
		collectOwnProperties(child, properties)
	}
	return properties
}

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
	if nil == cls.Children {
		return bson.M{"type": cm}
	}
	ar := make([]interface{}, 0, len(cls.Children))
	ar = append(ar, cm)
	for _, child := range cls.Children {
		ar = append(ar, stringutils.Underscore(child.Name))
	}
	return bson.M{"type": bson.M{"$in": ar}}
}
func buildQueryStatement(cls *ClassDefinition, params map[string]string) (bson.M, error) {
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
	return q, nil
}

func (self *mdb_server) FindBy(cls *ClassDefinition, params map[string]string) ([]map[string]interface{}, error) {
	s, err := buildQueryStatement(cls, params)
	if nil != err {
		return nil, err
	}
	q := self.session.C(cls.CollectionName()).Find(s)
	if nil == q {
		return nil, errors.New("return nil result")
	}
	offset := params["offset"]
	if "" != offset {
		n, err := strconv.Atoi(offset)
		if nil != err {
			return nil, errors.New("'offset' is not a integer - " + offset)
		}
		q.Skip(n)
	}
	limit := params["limit"]
	if "" != limit {
		n, err := strconv.Atoi(limit)
		if nil != err {
			return nil, errors.New("'limit' is not a integer - " + limit)
		}
		q.Limit(n)
	}

	results := make([]map[string]interface{}, 0, 10)
	it := q.Iter()
	attributes := make(map[string]interface{})
	for it.Next(&attributes) {
		attributes, err = self.postRead(cls, attributes)
		if nil != err {
			return nil, err
		}
		results = append(results, attributes)
		attributes = make(map[string]interface{})
	}

	err = it.Err()
	if nil != err {
		return nil, err
	}

	return results, nil
}

func (self *mdb_server) Count(cls *ClassDefinition, params map[string]string) (int, error) {
	s, err := buildQueryStatement(cls, params)
	if nil != err {
		return -1, err
	}
	q := self.session.C(cls.CollectionName()).Find(s)
	if nil == q {
		return -1, errors.New("return nil result")
	}
	return q.Count()
}
