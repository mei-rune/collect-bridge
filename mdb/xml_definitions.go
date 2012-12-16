package mdb

import (
	"encoding/xml"
)

type XMLClassDefinitions struct {
	XMLName      xml.Name             `xml:"http://schemas.meijing.com.cn/mdbs/1/typeDefinitions classDefinitions"`
	LastModified string               `xml:"lastModified,attr"`
	Definitions  []XMLClassDefinition `xml:"class"`
}

type XMLClassDefinition struct {
	Name string `xml:"name,attr"`
	//DisplayName string        `xml:"displayName,attr,omitempty"`
	Base                string                   `xml:"base,attr,omitempty"`
	Properties          []XMLPropertyDefinition  `xml:"property"`
	BelongsTo           []XMLBelongsTo           `xml:"belongs_to"`
	HasMany             []XMLHasMang             `xml:"has_many"`
	HasAndBelongsToMany []XMLHasAndBelongsToMany `xml:"has_and_belongs_to_many"`
}

type XMLBelongsTo struct {
	Name   string `xml:"name,attr,omitempty"`
	Target string `xml:",chardata"`
}

type XMLHasMang struct {
	//XMLName struct{} `xml:"has_many"`
	//Name   string `xml:"name,attr,omitempty"`
	Target string `xml:",chardata"`
}

type XMLHasAndBelongsToMany struct {
	//XMLName struct{} `xml:"has_and_belongs_to_many"`
	Target string `xml:",chardata"`
}

// <xs:element name="ref" minOccurs="0" maxOccurs="unbounded">
// 	<xs:annotation>
// 		<xs:documentation>指向一个XMLClassDefinition作为它的父类</xs:documentation>
// 	</xs:annotation>
// 	<xs:complexType>
// 		<xs:attribute name="id" type="xs:IDREF" use="required">
// 			<xs:annotation>
// 				<xs:documentation>其值应是某个XMLClassDefinition实例的name值</xs:documentation>
// 			</xs:annotation>
// 		</xs:attribute>
// 	</xs:complexType>
// </xs:element>
// type XMLRefDefinition struct {
// 	XMLName xml.Name `xml:"ref"`
// 	Name    string   `xml:"id,attr"`
// }

type XMLPropertyDefinition struct {
	//XMLName xml.Name `xml:"property"`
	Name string `xml:"name,attr"`
	//DisplayName string        `xml:"displayName,attr,omitempty"`
	Restrictions XMLRestrictionsDefinition
}

type XMLRestrictionsDefinition struct {
	XMLName xml.Name `xml:"restriction"`
	Type    string   `xml:"base,attr"`

	DefaultValue string    `xml:"defaultValue,omitempty"`
	Enumerations *[]string `xml:"enumeration>value,omitempty"`
	Pattern      string    `xml:"pattern,omitempty"`
	MinValue     string    `xml:"minValue,omitempty"`
	MaxValue     string    `xml:"maxValue,omitempty"`
	Length       string    `xml:"Length,omitempty"`
	MinLength    string    `xml:"minLength,omitempty"`
	MaxLength    string    `xml:"maxLength,omitempty"`

	// DefaultValue DefaultValueRestrictionDefinition `xml:"defaultValue,omitempty"`
	// Enumerations XMLEnumerationRestrictionDefinition  `xml:",omitempty"`
	// Pattern      PatternRestrictionDefinition      `xml:"pattern,omitempty"`
	// MinValue     MinValueRestrictionDefinition     `xml:"minValue,omitempty"`
	// MaxValue     MaxValueRestrictionDefinition     `xml:"maxValue,omitempty"`
	// Length       LengthRestrictionDefinition       `xml:"Length,omitempty"`
	// MinLength    MinLengthRestrictionDefinition    `xml:"minLength,omitempty"`
	// MaxLength    MaxLengthRestrictionDefinition    `xml:"maxLength,omitempty"`
}

// type DefaultValueRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"defaultValue"`
// 	Value   string   `xml:"value"`
// }

// type XMLEnumerationValue struct {
// 	XMLName xml.Name `xml:"value"`
// 	Value   string   `xml:",chardata"`
// }

// type XMLEnumerationRestrictionDefinition struct {
// 	//XMLName xml.Name `xml:"value"`
// 	Value []string `xml:"value,omitempty"`
// }

// type PatternRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"pattern"`
// 	Value   string   `xml:"value,chardata"`
// }

// type MinValueRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"minValue"`
// 	Value   string   `xml:"value, chardata"`
// }

// type MaxValueRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"maxValue"`
// 	Value   string   `xml:"value, chardata"`
// }

// type LengthRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"length"`
// 	Value   string   `xml:"value, chardata"`
// }

// type MinLengthRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"minLength"`
// 	Value   string   `xml:"value, chardata"`
// }

// type MaxLengthRestrictionDefinition struct {
// 	XMLName xml.Name `xml:"maxLength"`
// 	Value   string   `xml:"value, chardata"`
// }
