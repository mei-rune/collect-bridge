package as

import (
	"reflect"
	"testing"
)

func testAsMap(t *testing.T, v interface{}, excepted map[string]interface{}) {

	i8, err := AsMap(v)
	if nil != err {
		t.Errorf("%v to int8 failed, excepted is %d", v, excepted)
	}

	if !reflect.DeepEqual(excepted, i8) {
		t.Errorf("%v to int8 failed, excepted is %v, actual is %v", v, excepted, i8)
	}
}
func testAsInt8(t *testing.T, v interface{}, excepted int8) {

	i8, err := AsInt8(v)
	if nil != err {
		t.Errorf("%v to int8 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to int8 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsInt8Failed(t *testing.T, v interface{}) {
	_, err := AsInt8(v)
	if nil == err {
		t.Errorf("%v to int8 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsInt16(t *testing.T, v interface{}, excepted int16) {

	i8, err := AsInt16(v)
	if nil != err {
		t.Errorf("%v to int16 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to int16 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsInt16Failed(t *testing.T, v interface{}) {
	_, err := AsInt16(v)
	if nil == err {
		t.Errorf("%v to int16 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsInt32(t *testing.T, v interface{}, excepted int32) {

	i8, err := AsInt32(v)
	if nil != err {
		t.Errorf("%v to int32 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to int32 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsInt32Failed(t *testing.T, v interface{}) {
	_, err := AsInt32(v)
	if nil == err {
		t.Errorf("%v to int32 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsInt64(t *testing.T, v interface{}, excepted int64) {

	i8, err := AsInt64(v)
	if nil != err {
		t.Errorf("%v to int64 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to int64 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsInt64Failed(t *testing.T, v interface{}) {
	_, err := AsInt64(v)
	if nil == err {
		t.Errorf("%v to int64 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsUint8(t *testing.T, v interface{}, excepted uint8) {

	i8, err := AsUint8(v)
	if nil != err {
		t.Errorf("%v to uint8 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to uint8 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsUint8Failed(t *testing.T, v interface{}) {
	_, err := AsUint8(v)
	if nil == err {
		t.Errorf("%v to uint8 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsUint16(t *testing.T, v interface{}, excepted uint16) {

	i8, err := AsUint16(v)
	if nil != err {
		t.Errorf("%v to uint16 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to uint16 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsUint16Failed(t *testing.T, v interface{}) {
	_, err := AsUint16(v)
	if nil == err {
		t.Errorf("%v to uint16 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsUint32(t *testing.T, v interface{}, excepted uint32) {

	i8, err := AsUint32(v)
	if nil != err {
		t.Errorf("%v to uint32 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to uint32 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsUint32Failed(t *testing.T, v interface{}) {
	_, err := AsUint32(v)
	if nil == err {
		t.Errorf("%v to uint32 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsUint64(t *testing.T, v interface{}, excepted uint64) {

	i8, err := AsUint64(v)
	if nil != err {
		t.Errorf("%v to uint64 failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to uint64 failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func testAsUint64Failed(t *testing.T, v interface{}) {
	_, err := AsUint64(v)
	if nil == err {
		t.Errorf("%v to uint64 failed, excepted throw a error, actual return ok", v)
	}
}

func testAsString(t *testing.T, v interface{}, excepted string) {

	i8, err := AsString(v)
	if nil != err {
		t.Errorf("%v to string failed, excepted is %d", v, excepted)
	}

	if excepted != i8 {
		t.Errorf("%v to string failed, excepted is %d, actual is %d", v, excepted, i8)
	}
}

func TestAs(t *testing.T) {
	testAsMap(t, "12", nil)

	testAsInt8(t, "12", 12)
	testAsInt8(t, int64(12), 12)
	testAsInt8(t, int32(12), 12)
	testAsInt8(t, int16(12), 12)
	testAsInt8(t, int8(12), 12)
	testAsInt8(t, int(12), 12)

	testAsInt8(t, uint64(0), 0)
	testAsInt8(t, uint32(0), 0)
	testAsInt8(t, uint16(0), 0)
	testAsInt8(t, uint8(0), 0)
	testAsInt8(t, uint(0), 0)

	testAsInt8Failed(t, "128")
	testAsInt8Failed(t, int64(128))
	testAsInt8Failed(t, int32(128))
	testAsInt8Failed(t, int16(128))
	testAsInt8Failed(t, int(128))

	testAsInt8Failed(t, uint64(128))
	testAsInt8Failed(t, uint32(128))
	testAsInt8Failed(t, uint16(128))
	testAsInt8Failed(t, uint8(128))
	testAsInt8Failed(t, uint(128))

	testAsInt16(t, "12", 12)
	testAsInt16(t, int64(12), 12)
	testAsInt16(t, int32(12), 12)
	testAsInt16(t, int16(12), 12)
	testAsInt16(t, int8(12), 12)
	testAsInt16(t, int(12), 12)

	testAsInt16(t, uint64(0), 0)
	testAsInt16(t, uint32(0), 0)
	testAsInt16(t, uint16(0), 0)
	testAsInt16(t, uint8(0), 0)
	testAsInt16(t, uint(0), 0)

	testAsInt16(t, "-32768", -32768)
	testAsInt16(t, int64(-32768), -32768)
	testAsInt16(t, int32(-32768), -32768)
	testAsInt16(t, int16(-32768), -32768)
	testAsInt16(t, int(-32768), -32768)

	testAsInt16(t, "32767", 32767)
	testAsInt16(t, int64(32767), 32767)
	testAsInt16(t, int32(32767), 32767)
	testAsInt16(t, int16(32767), 32767)
	testAsInt16(t, int(32767), 32767)

	testAsInt16Failed(t, "-32769")
	testAsInt16Failed(t, int64(-32769))
	testAsInt16Failed(t, int32(-32769))
	testAsInt16Failed(t, int(-32769))

	testAsInt16Failed(t, "32768")
	testAsInt16Failed(t, int64(32768))
	testAsInt16Failed(t, int32(32768))
	testAsInt16Failed(t, int(32768))

	testAsInt16Failed(t, uint64(32768))
	testAsInt16Failed(t, uint32(32768))
	testAsInt16Failed(t, uint16(32768))
	testAsInt16Failed(t, uint(32768))

	testAsInt32(t, "12", 12)
	testAsInt32(t, int64(12), 12)
	testAsInt32(t, int32(12), 12)
	testAsInt32(t, int16(12), 12)
	testAsInt32(t, int8(12), 12)
	testAsInt32(t, int(12), 12)

	testAsInt32(t, uint64(0), 0)
	testAsInt32(t, uint32(0), 0)
	testAsInt32(t, uint16(0), 0)
	testAsInt32(t, uint8(0), 0)
	testAsInt32(t, uint(0), 0)

	testAsInt32(t, "-32768", -32768)
	testAsInt32(t, int64(-32768), -32768)
	testAsInt32(t, int32(-32768), -32768)
	testAsInt32(t, int(-32768), -32768)

	testAsInt32(t, "2147483647", 2147483647)
	testAsInt32(t, int64(2147483647), 2147483647)
	testAsInt32(t, int32(2147483647), 2147483647)
	testAsInt32(t, int(2147483647), 2147483647)

	testAsInt32(t, uint64(2147483647), 2147483647)
	testAsInt32(t, uint32(2147483647), 2147483647)
	testAsInt32(t, uint(2147483647), 2147483647)

	testAsInt32Failed(t, "-2147483649")
	testAsInt32Failed(t, int64(-2147483649))

	testAsInt32Failed(t, uint32(2147483648))
	testAsInt32Failed(t, uint64(2147483648))

	testAsInt32Failed(t, uint64(2147483648))
	testAsInt32Failed(t, uint32(2147483648))
	testAsInt32Failed(t, uint(2147483648))

	testAsUint8(t, "12", 12)
	testAsUint8(t, uint64(12), 12)
	testAsUint8(t, uint32(12), 12)
	testAsUint8(t, uint16(12), 12)
	testAsUint8(t, uint8(12), 12)
	testAsUint8(t, uint(12), 12)

	testAsUint8(t, int64(12), 12)
	testAsUint8(t, int32(12), 12)
	testAsUint8(t, int16(12), 12)
	testAsUint8(t, int8(12), 12)
	testAsUint8(t, int(12), 12)

	testAsUint8(t, "0", 0)
	testAsUint8(t, uint64(0), 0)
	testAsUint8(t, uint32(0), 0)
	testAsUint8(t, uint16(0), 0)
	testAsUint8(t, uint8(0), 0)
	testAsUint8(t, uint(0), 0)

	testAsUint8(t, "255", 255)
	testAsUint8(t, uint64(255), 255)
	testAsUint8(t, uint32(255), 255)
	testAsUint8(t, uint16(255), 255)
	testAsUint8(t, uint8(255), 255)
	testAsUint8(t, uint(255), 255)

	testAsUint8Failed(t, "256")
	testAsUint8Failed(t, uint64(256))
	testAsUint8Failed(t, uint32(256))
	testAsUint8Failed(t, uint16(256))
	testAsUint8Failed(t, uint(256))

	testAsUint8Failed(t, int64(-12))
	testAsUint8Failed(t, int32(-12))
	testAsUint8Failed(t, int16(-12))
	testAsUint8Failed(t, int8(-12))
	testAsUint8Failed(t, int(-12))

	testAsUint16(t, "12", 12)
	testAsUint16(t, uint64(12), 12)
	testAsUint16(t, uint32(12), 12)
	testAsUint16(t, uint16(12), 12)
	testAsUint16(t, uint8(12), 12)
	testAsUint16(t, uint(12), 12)

	testAsUint16(t, int64(12), 12)
	testAsUint16(t, int32(12), 12)
	testAsUint16(t, int16(12), 12)
	testAsUint16(t, int8(12), 12)
	testAsUint16(t, int(12), 12)

	testAsUint16(t, "0", 0)
	testAsUint16(t, uint64(0), 0)
	testAsUint16(t, uint32(0), 0)
	testAsUint16(t, uint16(0), 0)
	testAsUint16(t, uint8(0), 0)
	testAsUint16(t, uint(0), 0)

	testAsUint16(t, "65535", 65535)
	testAsUint16(t, uint64(65535), 65535)
	testAsUint16(t, uint32(65535), 65535)
	testAsUint16(t, uint(65535), 65535)

	testAsUint16Failed(t, "65536")
	testAsUint16Failed(t, uint64(65536))
	testAsUint16Failed(t, uint32(65536))
	testAsUint16Failed(t, uint(65536))

	testAsUint16Failed(t, int64(-12))
	testAsUint16Failed(t, int32(-12))
	testAsUint16Failed(t, int16(-12))
	testAsUint16Failed(t, int8(-12))
	testAsUint16Failed(t, int(-12))

	testAsUint32(t, "12", 12)
	testAsUint32(t, uint64(12), 12)
	testAsUint32(t, uint32(12), 12)
	testAsUint32(t, uint16(12), 12)
	testAsUint32(t, uint8(12), 12)
	testAsUint32(t, uint(12), 12)

	testAsUint32(t, int64(12), 12)
	testAsUint32(t, int32(12), 12)
	testAsUint32(t, int16(12), 12)
	testAsUint32(t, int8(12), 12)
	testAsUint32(t, int(12), 12)

	testAsUint32(t, "0", 0)
	testAsUint32(t, uint64(0), 0)
	testAsUint32(t, uint32(0), 0)
	testAsUint32(t, uint16(0), 0)
	testAsUint32(t, uint8(0), 0)
	testAsUint32(t, uint(0), 0)

	testAsUint32(t, "4294967295", 4294967295)
	testAsUint32(t, uint64(4294967295), 4294967295)
	testAsUint32(t, uint32(4294967295), 4294967295)
	testAsUint32(t, uint(4294967295), 4294967295)

	testAsUint32Failed(t, "4294967296")
	testAsUint32Failed(t, uint64(4294967296))

	testAsUint32Failed(t, int64(-12))
	testAsUint32Failed(t, int32(-12))
	testAsUint32Failed(t, int16(-12))
	testAsUint32Failed(t, int8(-12))
	testAsUint32Failed(t, int(-12))

	testAsUint64(t, "12", 12)
	testAsUint64(t, uint64(12), 12)
	testAsUint64(t, uint32(12), 12)
	testAsUint64(t, uint16(12), 12)
	testAsUint64(t, uint8(12), 12)
	testAsUint64(t, uint(12), 12)

	testAsUint64(t, int64(12), 12)
	testAsUint64(t, int32(12), 12)
	testAsUint64(t, int16(12), 12)
	testAsUint64(t, int8(12), 12)
	testAsUint64(t, int(12), 12)

	testAsUint64(t, "0", 0)
	testAsUint64(t, uint64(0), 0)
	testAsUint64(t, uint32(0), 0)
	testAsUint64(t, uint16(0), 0)
	testAsUint64(t, uint8(0), 0)
	testAsUint64(t, uint(0), 0)

	testAsUint64(t, "18446744073709551615", 18446744073709551615)
	testAsUint64(t, uint64(18446744073709551615), 18446744073709551615)

	testAsUint64Failed(t, "18446744073709551616")

	testAsUint64Failed(t, int64(-12))
	testAsUint64Failed(t, int32(-12))
	testAsUint64Failed(t, int16(-12))
	testAsUint64Failed(t, int8(-12))
	testAsUint64Failed(t, int(-12))

	testAsString(t, "12", "12")
	testAsString(t, uint64(12), "12")
	testAsString(t, uint32(12), "12")
	testAsString(t, uint16(12), "12")
	testAsString(t, uint8(12), "12")
	testAsString(t, uint(12), "12")

	testAsString(t, int64(12), "12")
	testAsString(t, int32(12), "12")
	testAsString(t, int16(12), "12")
	testAsString(t, int8(12), "12")
	testAsString(t, int(12), "12")

}
