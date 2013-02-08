package snmp

import (
	"reflect"
	"testing"
)

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

func TestAs(t *testing.T) {

	testAsInt32(t, "12", 12)
	testAsInt32(t, int64(12), 12)
	testAsInt32(t, SnmpInt32(12), 12)
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
	testAsInt32(t, SnmpInt32(-32768), -32768)
	testAsInt32(t, int(-32768), -32768)

	testAsInt32(t, "2147483647", 2147483647)
	testAsInt32(t, int64(2147483647), 2147483647)
	testAsInt32(t, SnmpInt32(2147483647), 2147483647)
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

	testAsUint32(t, int64(12), 12)
	testAsUint32(t, SnmpInt32(12), 12)
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
	testAsUint32Failed(t, SnmpInt32(-12))
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
	testAsUint64(t, SnmpInt32(12), 12)
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

	testAsUint64Failed(t, SnmpInt32(-12))

}
