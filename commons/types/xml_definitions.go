package types

import (
	"encoding/xml"
)

type XMLClassDefinitions struct {
	XMLName      xml.Name             `xml:"http://schemas.meijing.com.cn/mdbs/1/typeDefinitions classDefinitions"`
	LastModified string               `xml:"lastModified,attr"`
	Definitions  []XMLClassDefinition `xml:"class"`
}

type XMLClassDefinition struct {
	Name                string                   `xml:"name,attr"`
	Base                string                   `xml:"base,attr,omitempty"`
	Attributes          []XMLAttributeDefinition `xml:"property"`
	BelongsTo           []XMLBelongsTo           `xml:"belongs_to"`
	HasMany             []XMLHasMany             `xml:"has_many"`
	HasOne              []XMLHasOne              `xml:"has_one"`
	HasAndBelongsToMany []XMLHasAndBelongsToMany `xml:"has_and_belongs_to_many"`
}

type XMLBelongsTo struct {
	Name   string `xml:"name,attr,omitempty"`
	Target string `xml:",chardata"`
}

type XMLHasMany struct {
	AttributeName string `xml:"attributeName,attr,omitempty"`
	ForeignKey    string `xml:"foreignKey,attr,omitempty"`
	Embedded      string `xml:"embedded,attr,omitempty"`
	Polymorphic   string `xml:"polymorphic,attr,omitempty"`
	Target        string `xml:",chardata"`
}

type XMLHasOne struct {
	AttributeName string `xml:"attributeName,attr,omitempty"`
	ForeignKey    string `xml:"foreignKey,attr,omitempty"`
	Embedded      string `xml:"embedded,attr,omitempty"`
	Target        string `xml:",chardata"`
}

type XMLHasAndBelongsToMany struct {
	ForeignKey string `xml:"foreignKey,attr,omitempty"`
	Through    string `xml:"through,attr,omitempty"`
	Target     string `xml:",chardata"`
}

type XMLAttributeDefinition struct {
	Name         string `xml:"name,attr"`
	Restrictions XMLRestrictionsDefinition
}

type XMLRestrictionsDefinition struct {
	XMLName    xml.Name `xml:"restriction"`
	Type       string   `xml:"base,attr"`
	Collection string   `xml:"collection,attr,omitempty"`

	ReadOnly     *XMLReadOnly `xml:",omitempty"`
	Unique       *XMLUnique   `xml:",omitempty"`
	Required     *XMLRequired `xml:",omitempty"`
	DefaultValue string       `xml:"defaultValue,omitempty"`
	Enumerations *[]string    `xml:"enumeration>value,omitempty"`
	Pattern      string       `xml:"pattern,omitempty"`
	MinValue     string       `xml:"minValue,omitempty"`
	MaxValue     string       `xml:"maxValue,omitempty"`
	Length       string       `xml:"length,omitempty"`
	MinLength    string       `xml:"minLength,omitempty"`
	MaxLength    string       `xml:"maxLength,omitempty"`
}

type XMLRequired struct {
	XMLName xml.Name `xml:"required"`
}

type XMLReadOnly struct {
	XMLName xml.Name `xml:"readonly"`
}

type XMLUnique struct {
	XMLName xml.Name `xml:"unique"`
}
