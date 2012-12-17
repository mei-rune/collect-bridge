package mdb

import (
	"regexp"
	"testing"
	"time"
)

func TestDate(t *testing.T) {

	v1, _ := time.Parse("2006-01-02 15:04:05", "2009-10-11 12:12:11")
	v2, _ := time.Parse("2006-01-02 15:04:05", "2009-10-11 12:12:12")
	v3, _ := time.Parse("2006-01-02 15:04:05", "2009-10-11 12:12:13")
	v4, _ := time.Parse("2006-01-02 15:04:05", "2009-10-11 12:12:14")
	v5, _ := time.Parse("2006-01-02 15:04:05", "2009-10-11 12:12:15")
	v6, _ := time.Parse("2006-01-02 15:04:05", "2009-10-11 12:12:16")

	value1 := SqlDateTime(v1)
	value2 := SqlDateTime(v2)
	value3 := SqlDateTime(v3)
	value4 := SqlDateTime(v4)
	value5 := SqlDateTime(v5)
	value6 := SqlDateTime(v6)

	var validator DateValidator

	assertTrue := func(value interface{}) {
		if ok, err := validator.Validate(value, nil); !ok {
			t.Errorf("test Date failed, %s", err.Error())
		}
	}
	assertFalse := func(value interface{}) {
		if ok, _ := validator.Validate(value, nil); ok {
			t.Errorf("test Date failed")
		}
	}

	assertTrue(value1)
	assertTrue(value5)
	assertTrue(value6)

	validator.HasMax = true
	validator.MaxValue = time.Time(value5)
	assertTrue(value1)
	assertTrue(value5)
	assertFalse(value6)

	validator.HasMax = false
	validator.MaxValue = time.Time(value5)
	validator.HasMin = true
	validator.MinValue = time.Time(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value6)

	validator.HasMax = true
	validator.MaxValue = time.Time(value5)
	validator.HasMin = true
	validator.MinValue = time.Time(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value3)
	assertTrue(value4)
	assertTrue(value5)
	assertFalse(value6)
}

func TestInteger(t *testing.T) {
	var value1 SqlInteger64 = SqlInteger64(1)
	var value2 SqlInteger64 = SqlInteger64(2)
	var value3 SqlInteger64 = SqlInteger64(3)
	var value4 SqlInteger64 = SqlInteger64(4)
	var value5 SqlInteger64 = SqlInteger64(5)
	var value6 SqlInteger64 = SqlInteger64(6)

	var validator IntegerValidator

	assertTrue := func(value interface{}) {
		if ok, err := validator.Validate(value, nil); !ok {
			t.Errorf("test integer failed, %s", err.Error())
		}
	}
	assertFalse := func(value interface{}) {
		if ok, _ := validator.Validate(value, nil); ok {
			t.Errorf("test integer failed")
		}
	}

	assertTrue(value1)
	assertTrue(value5)
	assertTrue(value6)

	validator.HasMax = true
	validator.MaxValue = int64(value5)
	assertTrue(value1)
	assertTrue(value5)
	assertFalse(value6)

	validator.HasMax = false
	validator.MaxValue = int64(value5)
	validator.HasMin = true
	validator.MinValue = int64(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value6)

	validator.HasMax = true
	validator.MaxValue = int64(value5)
	validator.HasMin = true
	validator.MinValue = int64(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value3)
	assertTrue(value4)
	assertTrue(value5)
	assertFalse(value6)
}

func TestDouble(t *testing.T) {

	var value1 SqlDecimal = 1.0
	var value2 SqlDecimal = 2.0
	var value3 SqlDecimal = 3.0
	var value4 SqlDecimal = 4.0
	var value5 SqlDecimal = 5.0
	var value6 SqlDecimal = 6.0

	var validator DecimalValidator

	assertTrue := func(value interface{}) {
		if ok, err := validator.Validate(value, nil); !ok {
			t.Errorf("test float failed, %s", err.Error())
		}
	}
	assertFalse := func(value interface{}) {
		if ok, _ := validator.Validate(value, nil); ok {
			t.Errorf("test float failed")
		}
	}

	assertTrue(value1)
	assertTrue(value5)
	assertTrue(value6)

	validator.HasMax = true
	validator.MaxValue = float64(value5)
	assertTrue(value1)
	assertTrue(value5)
	assertFalse(value6)

	validator.HasMax = false
	validator.MaxValue = float64(value5)
	validator.HasMin = true
	validator.MinValue = float64(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value6)

	validator.HasMax = true
	validator.MaxValue = float64(value5)
	validator.HasMin = true
	validator.MinValue = float64(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value3)
	assertTrue(value4)
	assertTrue(value5)
	assertFalse(value6)
}

func TestString(t *testing.T) {

	var value1 SqlString = "aaa1"
	var value2 SqlString = "aaaa2"
	var value3 SqlString = "aaaaa3"
	var value4 SqlString = "aaaaaa4"
	var value5 SqlString = "aaaaaaa5"
	var value6 SqlString = "aaaaaaaa6"

	var checker Validator
	var validator StringLengthValidator

	checker = &validator

	assertTrue := func(value interface{}) {
		if ok, err := checker.Validate(value, nil); !ok {
			t.Errorf("test string failed, %s", err.Error())
		}
	}
	assertFalse := func(value interface{}) {
		if ok, _ := checker.Validate(value, nil); ok {
			t.Errorf("test string failed")
		}
	}

	validator.MaxLength = -1
	validator.MinLength = -1
	assertTrue(value1)
	assertTrue(value5)
	assertTrue(value6)

	validator.MaxLength = len(value5)
	validator.MinLength = -1
	assertTrue(value1)
	assertTrue(value5)
	assertFalse(value6)

	validator.MaxLength = -1
	validator.MinLength = len(value2)

	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value6)

	validator.MaxLength = len(value5)
	validator.MinLength = len(value2)
	assertFalse(value1)
	assertTrue(value2)
	assertTrue(value3)
	assertTrue(value4)
	assertTrue(value5)
	assertFalse(value6)

	validator.MaxLength = -1
	validator.MinLength = -1

	pv := &PatternValidator{}
	checker = pv
	pv.Pattern, _ = regexp.Compile("a.*")
	assertFalse("ddd")
	assertTrue("aaa")
}
