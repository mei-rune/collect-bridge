package mdb

type TypeDefinition interface {
}

type IntegerTypeDefinition struct {
}

type DecimalTypeDefinition struct {
}

type StringTypeDefinition struct {
}

type DateTimeTypeDefinition struct {
}

type IpAddressTypeDefinition struct {
}

type PhysicalAddressTypeDefinition struct {
}

type PasswordTypeDefinition struct {
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
	switch {
	case "integer":
		return integerType
	case "decimal":
		return decimalType
	case "string":
		return stringType
	case "dateTime":
		return datetimeType
	case "ipAddress":
		return ipAddressType
	case "physicalAddress":
		return physicalAddressType
	case "password":
		return passwordType
	}
	return nil
}
