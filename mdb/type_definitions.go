package mdb

type TypeDefinition interface {
	CreateEnumerationValidator(values []string) (Validator, error)
	CreatePatternValidator(pattern string) (Validator, error)
	CreateRangeValidator(minValue, maxValue string) (Validator, error)
	CreateLengthValidator(minLength, maxLength string) (Validator, error)
	ConvertFrom(v interface{}) (interface{}, error)
}

type IntegerTypeDefinition struct {
}

func (self *IntegerTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *IntegerTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *IntegerTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *IntegerTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *IntegerTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

type DecimalTypeDefinition struct {
}

func (self *DecimalTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *DecimalTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *DecimalTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *DecimalTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *DecimalTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

type StringTypeDefinition struct {
}

func (self *StringTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *StringTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *StringTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *StringTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *StringTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

type DateTimeTypeDefinition struct {
}

func (self *DateTimeTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *DateTimeTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *DateTimeTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *DateTimeTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *DateTimeTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

type IpAddressTypeDefinition struct {
}

func (self *IpAddressTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *IpAddressTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *IpAddressTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *IpAddressTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *IpAddressTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

type PhysicalAddressTypeDefinition struct {
}

func (self *PhysicalAddressTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *PhysicalAddressTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *PhysicalAddressTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *PhysicalAddressTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *PhysicalAddressTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

type PasswordTypeDefinition struct {
}

func (self *PasswordTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not implemented")
}
func (self *PasswordTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not implemented")
}
func (self *PasswordTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not implemented")
}
func (self *PasswordTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not implemented")
}
func (self *PasswordTypeDefinition) ConvertFrom(v interface{}) (interface{}, error) {
	panic("not implemented")
}

var (
	integerType         IntegerTypeDefinition
	decimalType         DecimalTypeDefinition
	stringType          StringTypeDefinition
	datetimeType        DateTimeTypeDefinition
	ipAddressType       IpAddressTypeDefinition
	physicalAddressType PhysicalAddressTypeDefinition
	passwordType        PasswordTypeDefinition
)

func GetTypeDefinition(t string) TypeDefinition {
	switch t {
	case "integer":
		return &integerType
	case "decimal":
		return &decimalType
	case "string":
		return &stringType
	case "dateTime":
		return &datetimeType
	case "ipAddress":
		return &ipAddressType
	case "physicalAddress":
		return &physicalAddressType
	case "password":
		return &passwordType
	}
	return nil
}
