package mdb

import (
	a "commons/assert"
	"encoding/xml"
	//"fmt"
	"io/ioutil"
	//"os"
	"testing"
)

func TestOutXML(t *testing.T) {

	classes := &XMLClassDefinitions{
		LastModified: "123",
		Definitions:  make([]XMLClassDefinition, 0)}
	cl1 := &XMLClassDefinition{Name: "Person", Properties: []XMLPropertyDefinition{
		XMLPropertyDefinition{Name: "Id", Restrictions: XMLRestrictionsDefinition{Type: "integer", DefaultValue: "12"}},
		XMLPropertyDefinition{Name: "Sex", Restrictions: XMLRestrictionsDefinition{Type: "string", Enumerations: &[]string{"male", "female"}}},
		XMLPropertyDefinition{Name: "Name", Restrictions: XMLRestrictionsDefinition{Type: "string", Pattern: "a.*"}},
		XMLPropertyDefinition{Name: "Age", Restrictions: XMLRestrictionsDefinition{Type: "integer", MinValue: "1", MaxValue: "130"}},
		XMLPropertyDefinition{Name: "Address", Restrictions: XMLRestrictionsDefinition{Type: "string", MinLength: "10", MaxLength: "20"}}}}
	cl2 := &XMLClassDefinition{Name: "Employee", Base: "Person",
		BelongsTo:           []XMLBelongsTo{XMLBelongsTo{Name: "cc_id", Target: "CC"}, XMLBelongsTo{Name: "bb_id", Target: "BB"}},
		HasMany:             []XMLHasMany{XMLHasMany{Target: "DD"}, XMLHasMany{Target: "BB"}},
		HasAndBelongsToMany: []XMLHasAndBelongsToMany{XMLHasAndBelongsToMany{Target: "DD"}, XMLHasAndBelongsToMany{Target: "BB"}},
		Properties: []XMLPropertyDefinition{
			XMLPropertyDefinition{Name: "Id2", Restrictions: XMLRestrictionsDefinition{Type: "string", DefaultValue: "12"}},
			XMLPropertyDefinition{Name: "Sex2", Restrictions: XMLRestrictionsDefinition{Type: "string", Enumerations: &[]string{"male", "female"}}},
			XMLPropertyDefinition{Name: "Name2", Restrictions: XMLRestrictionsDefinition{Type: "string", Pattern: "a.*"}},
			XMLPropertyDefinition{Name: "Age2", Restrictions: XMLRestrictionsDefinition{Type: "string", MinValue: "1", MaxValue: "130"}},
			XMLPropertyDefinition{Name: "Address2", Restrictions: XMLRestrictionsDefinition{Type: "string", MinLength: "10", MaxLength: "20"}}}}

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
	a.Assert(t, len(person.Properties), a.Equals, 5, a.Commentf("check len of Properties"))
	a.Check(t, person.Properties[0].Name, a.Equals, "Id", a.Commentf("check name of Properties[0]"))

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

func TestReadXML(t *testing.T) {
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
	if 3 != len(xmlDefinitions.Definitions) {
		t.Errorf("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
		return
	}

	employee := xmlDefinitions.Definitions[0]
	person := xmlDefinitions.Definitions[1]
	company := xmlDefinitions.Definitions[2]

	a.Check(t, person.Name, a.Equals, "Person", a.Commentf("check Class name"))
	a.Check(t, person.Base, a.Equals, "", a.Commentf("check Base name"))
	a.Assert(t, len(person.Properties), a.Equals, 10, a.Commentf("check len of Properties"))
	a.Check(t, person.Properties[0].Name, a.Equals, "ID1", a.Commentf("check name of Properties[0]"))

	assertProperty := func(p1, p2 *XMLPropertyDefinition, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of properties[%d]", comment))
		a.Check(t, p1.Restrictions.Type, a.Equals, p2.Restrictions.Type,
			a.Commentf("check Restrictions.Type of properties[%d]", comment))
		a.Check(t, p1.Restrictions.DefaultValue, a.DeepEquals, p1.Restrictions.DefaultValue,
			a.Commentf("check Restrictions.DefaultValue of properties[%d]", comment))

		if !checkArray(p1.Restrictions.Enumerations, p2.Restrictions.Enumerations) {
			t.Errorf("check Restrictions.Enumerations properties[%d] failed, value is %v", comment, p1.Restrictions.Enumerations)
		}

		a.Check(t, p1.Restrictions.Length, a.Equals, p1.Restrictions.Length,
			a.Commentf("check Restrictions.Length properties[%d]", comment))
		a.Check(t, p1.Restrictions.MaxLength, a.Equals, p1.Restrictions.MaxLength,
			a.Commentf("check Restrictions.MaxLength properties[%d]", comment))
		a.Check(t, p1.Restrictions.MinLength, a.Equals, p1.Restrictions.MinLength,
			a.Commentf("check Restrictions.MinLength properties[%d]", comment))
		a.Check(t, p1.Restrictions.MaxValue, a.Equals, p1.Restrictions.MaxValue,
			a.Commentf("check Restrictions.MaxValue properties[%d]", comment))
		a.Check(t, p1.Restrictions.MinValue, a.Equals, p1.Restrictions.MinValue,
			a.Commentf("check Restrictions.MinValue properties[%d]", comment))
		a.Check(t, p1.Restrictions.Pattern, a.Equals, p1.Restrictions.Pattern,
			a.Commentf("check Restrictions.Pattern properties[%d]", comment))
	}

	assertBelongsTo := func(p1, p2 *XMLBelongsTo, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of belongs_to[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of belongs_to[%d]", comment))
	}

	assertHasMany := func(p1, p2 *XMLHasMany, comment int) {
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of has_many[%d]", comment))
	}

	assertProperty(&person.Properties[0], &XMLPropertyDefinition{Name: "ID1",
		Restrictions: XMLRestrictionsDefinition{Type: "integer", DefaultValue: "0"}}, 0)
	assertProperty(&person.Properties[1], &XMLPropertyDefinition{Name: "Name",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "mfk",
			MinLength:    "3", MaxLength: "13"}}, 1)
	assertProperty(&person.Properties[2], &XMLPropertyDefinition{Name: "Name2",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "mfk",
			Length:       "3"}}, 2)
	assertProperty(&person.Properties[3], &XMLPropertyDefinition{Name: "Age",
		Restrictions: XMLRestrictionsDefinition{Type: "integer",
			DefaultValue: "mfk",
			MinValue:     "3",
			MaxValue:     "313"}}, 3)
	assertProperty(&person.Properties[4], &XMLPropertyDefinition{Name: "Day",
		Restrictions: XMLRestrictionsDefinition{Type: "datetime",
			DefaultValue: "2009-12-12 12:23:23",
			MinValue:     "2009-12-11 10:23:23",
			MaxValue:     "2009-12-13 12:23:23"}}, 4)
	assertProperty(&person.Properties[5], &XMLPropertyDefinition{Name: "Mony",
		Restrictions: XMLRestrictionsDefinition{Type: "decimal",
			DefaultValue: "1.3",
			MinValue:     "1.0",
			MaxValue:     "3.0"}}, 5)
	assertProperty(&person.Properties[6], &XMLPropertyDefinition{Name: "IP",
		Restrictions: XMLRestrictionsDefinition{Type: "ipAddress",
			DefaultValue: "12.12.12.12"}}, 6)
	assertProperty(&person.Properties[7], &XMLPropertyDefinition{Name: "MAC",
		Restrictions: XMLRestrictionsDefinition{Type: "physicalAddress",
			DefaultValue: "12-12-12-12-12-12"}}, 7)
	assertProperty(&person.Properties[8], &XMLPropertyDefinition{Name: "Sex",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "male",
			Enumerations: &[]string{"male", "female"}}}, 8)
	assertProperty(&person.Properties[9], &XMLPropertyDefinition{Name: "Password",
		Restrictions: XMLRestrictionsDefinition{Type: "password",
			DefaultValue: "mfk"}}, 9)

	a.Check(t, employee.Name, a.Equals, "Employee", a.Commentf("check Class name"))
	a.Check(t, employee.Base, a.Equals, "Person", a.Commentf("check Base name"))

	a.Assert(t, len(employee.Properties), a.Equals, 2, a.Commentf("check len of Properties"))

	assertProperty(&employee.Properties[0], &XMLPropertyDefinition{Name: "Job",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "developer"}}, 0)
	assertProperty(&employee.Properties[1], &XMLPropertyDefinition{Name: "company_id",
		Restrictions: XMLRestrictionsDefinition{Type: "string"}}, 0)

	a.Check(t, company.Name, a.Equals, "Company", a.Commentf("check Class name"))

	a.Assert(t, len(company.Properties), a.Equals, 1, a.Commentf("check len of Properties"))

	assertProperty(&company.Properties[0], &XMLPropertyDefinition{Name: "Name",
		Restrictions: XMLRestrictionsDefinition{Type: "string",
			DefaultValue: "Sina"}}, 0)

	// if 3 != len(xmlDefinitions.Definitions) {
	// 	t.Errorf("", len(xmlDefinitions.Definitions))
	// 	return
	// }
	assertBelongsTo(&employee.BelongsTo[0], &XMLBelongsTo{Target: "Company", Name: "company_id"}, 0)
	assertHasMany(&company.HasMany[0], &XMLHasMany{Target: "Employee"}, 0)

}
