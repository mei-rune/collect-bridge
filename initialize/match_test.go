package main

import (
	"testing"
)

func TestMathc(t *testing.T) {
	if "listen_addresses = '*'   # what IP address(es) to listen on;" != replaceAddressAndPort("#listen_addresses = 'localhost'   # what IP address(es) to listen on;", "*", "80") {
		t.Error("address match failed")
	}

	if "listen_addressesa = '*'   # what IP address(es) to listen on;" == replaceAddressAndPort("#listen_addresses = 'localhost'   # what IP address(es) to listen on;", "*", "80") {
		t.Error("address match failed")
	}

	if "port = 80       # (change requires restart)" != replaceAddressAndPort("#port = 5432       # (change requires restart)", "*", "80") {
		t.Error("address match failed")
	}

	if "porta = 80       # (change requires restart)" == replaceAddressAndPort("#port = 5432       # (change requires restart)", "*", "80") {
		t.Error("address match failed")
	}
}
