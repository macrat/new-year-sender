package main

import (
	"strings"
	"testing"
)

func TestDateTime(t *testing.T) {
	date := DateTime{}

	if err := (&date).UnmarshalText([]byte("2018-04-03 14:01")); err != nil {
		t.Fatalf("Unmarshal: failed to parse: %s", err.Error())
	}

	if formatted := date.Format("2006/01/02T15:04"); formatted != "2018/04/03T14:01" {
		t.Errorf("Format: excepted %s but got %s", "2018/04/03T14:01", formatted)
	}

	if formatted := date.Format("2006/01/02T15:04"); strings.HasPrefix(formatted, "2018/04/03T14:01 ") {
		t.Errorf("Format: excepted starts with 2018/04/03T14:01 but got %s", formatted)
	}

	if bytes, err := date.MarshalText(); err != nil {
		t.Errorf("Marshal: failed marshal: %s", err.Error())
	} else if string(bytes) != "2018-04-03 14:01" {
		t.Errorf("Marshal: excepted %s but got %s", "2018-04-03 14:01", string(bytes))
	}
}

func TestAddress(t *testing.T) {
	addr := Address{}

	if err := (&addr).UnmarshalText([]byte("hoge <fuga@example.com>")); err != nil {
		t.Fatalf("Unmarshal: failed to parse: %s", err.Error())
	}

	if str := addr.String(); str != "\"hoge\" <fuga@example.com>" {
		t.Errorf("Format: excepted %s but got %s", "\"hoge\" <fuga@example.com>", str)
	}

	if bytes, err := addr.MarshalText(); err != nil {
		t.Errorf("Marshal: failed marshal: %s", err.Error())
	} else if string(bytes) != "\"hoge\" <fuga@example.com>" {
		t.Errorf("Marshal: excepted %s but got %s", "\"hoge\" <fuga@example.com>", string(bytes))
	}

	if addr.IsEmpty() {
		t.Errorf("IsEmpty: excepted false but got true")
	}

	if !(Address{}).IsEmpty() {
		t.Errorf("IsEmpty: excepted true but got false")
	}

	if (Address{Address: "hoge@example.com"}).IsEmpty() {
		t.Errorf("IsEmpty: excepted false but got true")
	}

	if !(Address{Name: "fuga"}).IsEmpty() {
		t.Errorf("IsEmpty: excepted true but got false")
	}
}
