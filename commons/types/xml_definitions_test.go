package types

import (
	a "commons/assert"
	"encoding/xml"
	"io/ioutil"
	"testing"
)

func TestOutXML(t *testing.T) {
	classes := &XMLClassDefinitions{
		LastModified: "123",
		Definitions:  make([]XMLClassDefinition, 0)}
	cl1 := &XMLClassDefinition{Name: "Person", Attributes: []XMLAttributeDefinition{
		XMLAttributeDefinition{Name: "Id", Restrictions: XMLRestrictionsDefinition{Type: "integer", DefaultValue: "12"}},
		XMLAttributeDefinition{Name: "Sex", Restrictions: XMLRestrictionsDefinition{Type: "string", Enumerations: &[]string{"male", "female"}}},
		XMLAttributeDefinition{Name: "Name", Restrictions: XMLRestrictionsDefinition{Type: "string", Pattern: "a.*"}},
		XMLAttributeDefinition{Name: "Age", Restrictions: XMLRestrictionsDefinition{Type: "integer", MinValue: "1", MaxValue: "130"}},
		XMLAttributeDefinition{Name: "Address", Restrictions: XMLRestrictionsDefinition{Type: "string", MinLength: "10", MaxLength: "20"}}}}
	cl2 := &XMLClassDefinition{Name: "Employee", Base: "Person",
		BelongsTo:           []XMLBelongsTo{XMLBelongsTo{Name: "cc_id", Target: "CC"}, XMLBelongsTo{Name: "bb_id", Target: "BB"}},
		HasMany:             []XMLHasMany{XMLHasMany{Target: "DD"}, XMLHasMany{Target: "BB"}},
		HasOne:              []XMLHasOne{XMLHasOne{Target: "DD"}, XMLHasOne{Target: "BB"}},
		HasAndBelongsToMany: []XMLHasAndBelongsToMany{XMLHasAndBelongsToMany{Target: "DD"}, XMLHasAndBelongsToMany{Target: "BB"}},
		Attributes: []XMLAttributeDefinition{
			XMLAttributeDefinition{Name: "Id2", Restrictions: XMLRestrictionsDefinition{Type: "string", DefaultValue: "12"}},
			XMLAttributeDefinition{Name: "Sex2", Restrictions: XMLRestrictionsDefinition{Type: "string", Enumerations: &[]string{"male", "female"}}},
			XMLAttributeDefinition{Name: "Name2", Restrictions: XMLRestrictionsDefinition{Type: "string", Pattern: "a.*"}},
			XMLAttributeDefinition{Name: "Age2", Restrictions: XMLRestrictionsDefinition{Type: "string", MinValue: "1", MaxValue: "130"}},
			XMLAttributeDefinition{Name: "Address2", Restrictions: XMLRestrictionsDefinition{Type: "string", MinLength: "10", MaxLength: "20"}}}}

	classes.Definitions = append(classes.Definitions, *cl1, *cl2)

	output, err := xml.MarshalIndent(classes, "  ", "    ")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	//os.Stdout.Write([]byte(xml.Header))
	//os.Stdout.Write(output)

	var xmlDefinitions XMLClassDefinitions
	err = xml.Unmarshal(output, &xmlDefinitions)
	if nil != err {
		t.Errorf("unmarshal xml failed, %s", err.Error())
		return
	}
	if nil == xmlDefinitions.Definitions {
		t.Errorf("unmarshal xml error, classDefinition is nil")
		return
	}
	if 2 != len(xmlDefinitions.Definitions) {
		t.Errorf("unmarshal xml error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
		return
	}

	person := xmlDefinitions.Definitions[0]
	employee := xmlDefinitions.Definitions[1]

	a.Check(t, person.Name, a.Equals, "Person", a.Commentf("check Class name"))
	a.Check(t, person.Base, a.Equals, "", a.Commentf("check Base name"))
	a.Assert(t, len(person.Attributes), a.Equals, 5, a.Commentf("check len of Attributes"))
	a.Check(t, person.Attributes[0].Name, a.Equals, "Id", a.Commentf("check name of Attributes[0]"))

	a.Check(t, employee.Name, a.Equals, "Employee", a.Commentf("check Class name"))
	a.Check(t, employee.Base, a.Equals, "Person", a.Commentf("check Base name"))
}

// func TestXml(t *testing.T) {
// 	type Email struct {
// 		Where string `xml:"where,attr"`
// 		Addr  string
// 	}
// 	type Address struct {
// 		City, State string
// 	}
// 	type Result struct {
// 		XMLName xml.Name `xml:"Person"`
// 		Name    string   `xml:"FullName"`
// 		Phone   string
// 		Email   []Email  `xml:"email"`
// 		Groups  []string `xml:"Group>Value"`
// 		Address
// 	}
// 	v := Result{Name: "none", Phone: "none"}

// 	data := `
//     <Person>
//         <FullName>Grace R. Emlin</FullName>
//         <Company>Example Inc.</Company>
//         <email where="home">
//             <Addr>gre@example.com</Addr>
//         </email>
//         <email where='work'>
//             <Addr>gre@work.com</Addr>
//         </email>
//         <Group>
//             <Value>Friends</Value>
//             <Value>Squash</Value>
//         </Group>
//         <City>Hanga Roa</City>
//         <State>Easter Island</State>
//     </Person>
// `
// 	err := xml.Unmarshal([]byte(data), &v)
// 	if err != nil {
// 		fmt.Printf("error: %v", err)
// 		return
// 	}
// 	fmt.Printf("XMLName: %#v\n", v.XMLName)
// 	fmt.Printf("Name: %q\n", v.Name)
// 	fmt.Printf("Phone: %q\n", v.Phone)
// 	fmt.Printf("Email: %v\n", v.Email)
// 	fmt.Printf("Groups: %v\n", v.Groups)
// 	fmt.Printf("Address: %v\n", v.Address)
// }

func checkArray(s1, s2 *[]string) bool {

	if nil == s1 || nil == *s1 || 0 == len(*s1) {
		if nil == s2 || nil == *s2 || 0 == len(*s2) {
			return true
		}
		return false
	}

	if len(*s1) != len(*s2) {
		return false
	}

	for i, s := range *s1 {
		if s != (*s2)[i] {
			return false
		}
	}
	return true
}

// func TestXML2(t *testing.T) {
// 	bytes, err := ioutil.ReadFile("test/test_property_override.xml")
// 	if nil != err {
// 		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
// 		return
// 	}
// 	var xmlDefinitions XMLClassDefinitions
// 	err = xml.Unmarshal(bytes, &xmlDefinitions)
// 	if nil != err {
// 		t.Errorf("unmarshal xml 'test/test1.xml' failed, %s", err.Error())
// 		return
// 	}
// 	if nil == xmlDefinitions.Definitions {
// 		t.Errorf("unmarshal xml 'test/test1.xml' error, classDefinition is nil")
// 		return
// 	}
// 	if 3 != len(xmlDefinitions.Definitions) {
// 		t.Errorf("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
// 		return
// 	}

// 	output, err := xml.MarshalIndent(xmlDefinitions, "  ", "    ")
// 	if err != nil {
// 		t.Errorf("error: %v\n", err)
// 	}
// 	os.Stdout.Write([]byte(xml.Header))
// 	os.Stdout.Write(output)
// }

func TestXML1(t *testing.T) {
	bytes, err := ioutil.ReadFile("test/test1.xml")
	if nil != err {
		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
		return
	}
	var xmlDefinitions XMLClassDefinitions
	err = xml.Unmarshal(bytes, &xmlDefinitions)
	if nil != err {
		t.Errorf("unmarshal xml 'test/test1.xml' failed, %s", err.Error())
		return
	}
	if nil == xmlDefinitions.Definitions {
		t.Errorf("unmarshal xml 'test/test1.xml' error, classDefinition is nil")
		return
	}
	if 4 != len(xmlDefinitions.Definitions) {
		t.Errorf("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
		return
	}

	employee := xmlDefinitions.Definitions[0]
	boss := xmlDefinitions.Definitions[1]
	person := xmlDefinitions.Definitions[2]
	company := xmlDefinitions.Definitions[3]

	a.Check(t, person.Name, a.Equals, "Person", a.Commentf("check Class name of person"))
	a.Check(t, person.Base, a.Equals, "", a.Commentf("check Base name of person"))
	a.Assert(t, len(person.Attributes), a.Equals, 10, a.Commentf("check len of Attributes of person"))
	a.Check(t, person.Attributes[0].Name, a.Equals, "ID1", a.Commentf("check name of Attributes[0] of person"))

	assertProperty := func(p1, p2 *XMLAttributeDefinition, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of properties[%d]", comment))
		a.Check(t, p1.Restrictions.Type, a.Equals, p2.Restrictions.Type,
			a.Commentf("check Restrictions.Type of properties[%d]", comment))
		a.Check(t, p1.Restrictions.DefaultValue, a.DeepEquals, p2.Restrictions.DefaultValue,
			a.Commentf("check Restrictions.DefaultValue of properties[%d]", comment))

		a.Check(t, p1.Restrictions.Required, a.DeepEquals, p2.Restrictions.Required,
			a.Commentf("check Restrictions.Required of properties[%d]", comment))

		if !checkArray(p1.Restrictions.Enumerations, p2.Restrictions.Enumerations) {
			t.Errorf("check Restrictions.Enumerations properties[%d] failed, value is %v", comment, p1.Restrictions.Enumerations)
		}

		a.Check(t, p1.Restrictions.Collection, a.Equals, p2.Restrictions.Collection,
			a.Commentf("check Restrictions.Collection properties[%d]", comment))
		a.Check(t, p1.Restrictions.Length, a.Equals, p2.Restrictions.Length,
			a.Commentf("check Restrictions.Length properties[%d]", comment))
		a.Check(t, p1.Restrictions.MaxLength, a.Equals, p2.Restrictions.MaxLength,
			a.Commentf("check Restrictions.MaxLength properties[%d]", comment))
		a.Check(t, p1.Restrictions.MinLength, a.Equals, p2.Restrictions.MinLength,
			a.Commentf("check Restrictions.MinLength properties[%d]", comment))
		a.Check(t, p1.Restrictions.MaxValue, a.Equals, p2.Restrictions.MaxValue,
			a.Commentf("check Restrictions.MaxValue properties[%d]", comment))
		a.Check(t, p1.Restrictions.MinValue, a.Equals, p2.Restrictions.MinValue,
			a.Commentf("check Restrictions.MinValue properties[%d]", comment))
		a.Check(t, p1.Restrictions.Pattern, a.Equals, p2.Restrictions.Pattern,
			a.Commentf("check Restrictions.Pattern properties[%d]", comment))
	}

	assertBelongsTo := func(p1, p2 *XMLBelongsTo, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of belongs_to[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of belongs_to[%d]", comment))
	}

	assertHasMany := func(p1, p2 *XMLHasMany, comment int) {
		a.Check(t, p1.ForeignKey, a.Equals, p2.ForeignKey, a.Commentf("check ForeignKey of has_many[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of has_many[%d]", comment))
	}

	assertHasOne := func(p1, p2 *XMLHasOne, comment int) {
		a.Check(t, p1.AttributeName, a.Equals, p2.AttributeName, a.Commentf("check AttributeName of has_one[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of has_one[%d]", comment))
	}

	assertProperty(&person.Attributes[0], &XMLAttributeDefinition{Name: "ID1",
		Restrictions: XMLRestrictionsDefinition{Type: "integer", DefaultValue: "0"}}, 0)
	assertProperty(&person.Attributes[1], &XMLAttributeDefinition{Name: "Name",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "mfk",
			MinLength:    "3", MaxLength: "13"}}, 1)
	assertProperty(&person.Attributes[2], &XMLAttributeDefinition{Name: "Name2",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "mfk",
			Length:       "3"}}, 2)
	assertProperty(&person.Attributes[3], &XMLAttributeDefinition{Name: "Age",
		Restrictions: XMLRestrictionsDefinition{Type: "integer",
			DefaultValue: "123",
			MinValue:     "3",
			MaxValue:     "313"}}, 3)
	assertProperty(&person.Attributes[4], &XMLAttributeDefinition{Name: "Day",
		Restrictions: XMLRestrictionsDefinition{Type: "datetime",
			DefaultValue: "2009-12-12T12:23:23+08:00",
			MinValue:     "2009-12-11T10:23:23+08:00",
			MaxValue:     "2009-12-13T12:23:23+08:00"}}, 4)
	assertProperty(&person.Attributes[5], &XMLAttributeDefinition{Name: "Mony",
		Restrictions: XMLRestrictionsDefinition{Type: "decimal",
			DefaultValue: "1.3",
			MinValue:     "1.0",
			MaxValue:     "3.0"}}, 5)
	assertProperty(&person.Attributes[6], &XMLAttributeDefinition{Name: "IP",
		Restrictions: XMLRestrictionsDefinition{Type: "ipAddress",
			DefaultValue: "12.12.12.12"}}, 6)
	assertProperty(&person.Attributes[7], &XMLAttributeDefinition{Name: "MAC",
		Restrictions: XMLRestrictionsDefinition{Type: "physicalAddress",
			DefaultValue: "12-12-12-12-12-12"}}, 7)
	assertProperty(&person.Attributes[8], &XMLAttributeDefinition{Name: "Sex",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "male",
			Enumerations: &[]string{"male", "female"}}}, 8)
	assertProperty(&person.Attributes[9], &XMLAttributeDefinition{Name: "Password",
		Restrictions: XMLRestrictionsDefinition{Type: "password",
			DefaultValue: "mfk"}}, 9)

	a.Check(t, employee.Name, a.Equals, "Employee", a.Commentf("check Class name"))
	a.Check(t, employee.Base, a.Equals, "Person", a.Commentf("check Base name"))

	a.Assert(t, len(employee.Attributes), a.Equals, 2, a.Commentf("check len of Attributes"))

	assertProperty(&employee.Attributes[0], &XMLAttributeDefinition{Name: "Job",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			Required: &XMLRequired{XMLName: xml.Name{Space: "http://schemas.meijing.com.cn/mdbs/1/typeDefinitions", Local: "required"}}}}, 0)
	assertProperty(&employee.Attributes[1], &XMLAttributeDefinition{Name: "company_test_id",
		Restrictions: XMLRestrictionsDefinition{Type: "objectId"}}, 0)

	a.Check(t, boss.Name, a.Equals, "Boss", a.Commentf("check Class name of boss"))
	a.Check(t, boss.Base, a.Equals, "Employee", a.Commentf("check Base name of boss"))

	a.Assert(t, len(boss.Attributes), a.Equals, 1, a.Commentf("check len of Attributes"))

	assertProperty(&boss.Attributes[0], &XMLAttributeDefinition{Name: "Job",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "boss", MinLength: "3", MaxLength: "13"}}, 0)

	a.Check(t, company.Name, a.Equals, "Company", a.Commentf("check Class company.name"))

	a.Assert(t, len(company.Attributes), a.Equals, 1, a.Commentf("check len of company.Attributes"))

	assertProperty(&company.Attributes[0], &XMLAttributeDefinition{Name: "Name",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "Sina"}}, 0)

	// if 3 != len(xmlDefinitions.Definitions) {
	// 	t.Errorf("", len(xmlDefinitions.Definitions))
	// 	return
	// }
	assertBelongsTo(&employee.BelongsTo[0], &XMLBelongsTo{Target: "Company", Name: "company_test_id"}, 0)
	assertHasMany(&company.HasMany[0], &XMLHasMany{Target: "Employee", ForeignKey: "company_test_id"}, 0)

	assertHasOne(&company.HasOne[0], &XMLHasOne{Target: "Boss", AttributeName: "boss"}, 0)

}
