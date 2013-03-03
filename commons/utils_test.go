package commons

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tt, _ := ParseDuration("12")
	if 12*time.Second != tt {
		t.Errorf("assert '%s' != %d, actaul is %s", "12", 12, tt)
	}

	tt, _ = ParseDuration("12ms")
	if 12*time.Millisecond != tt {
		t.Errorf("assert '%s' != %d", "12", 12)
	}

	tt, _ = ParseDuration("12s")
	if 12*time.Second != tt {
		t.Errorf("assert '%s' != %d", "12", 12)
	}

	tt, _ = ParseDuration("12m")
	if 60*12*time.Second != tt {
		t.Errorf("assert '%s' != %d", "12s", 60*12)
	}
	_, err := ParseDuration("12mss")
	if nil == err {
		t.Errorf("except parse '12ms' failed, actual return ok")
	}
	_, err = ParseDuration("ms")
	if nil == err {
		t.Errorf("except parse '12ms' failed, actual return ok")
	}
}
