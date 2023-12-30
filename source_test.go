package main

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
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

func TestAddressList(t *testing.T) {
	al := AddressList{}
	if err := yaml.Unmarshal([]byte("- hoge@fuga.com\n- foo <bar@baz.com>\n"), &al); err != nil {
		t.Fatalf("Unmarshal: failed to parse: %s", err.Error())
	}

	if str := al[0].String(); str != "<hoge@fuga.com>" {
		t.Errorf("Excepted first email is %#v but got %#v", `<hoge@fuga.com>`, str)
	}

	if str := al[1].String(); str != `"foo" <bar@baz.com>` {
		t.Errorf("Excepted first email is %#v but got %#v", `"foo" <bar@baz.com>`, str)
	}

	if str := al.String(); str != `<hoge@fuga.com>, "foo" <bar@baz.com>` {
		t.Errorf("Format: excepted %#v but got %#v", `<hoge@fuga.com>, "foo" <bar@baz.com>`, str)
	}
}

func TestTemplate(t *testing.T) {
	var tmpl Template
	if err := yaml.Unmarshal([]byte("Hello! {{.Text}} Best regards,"), &tmpl); err != nil {
		t.Fatalf("Unmarshal: failed to parse: %s", err.Error())
	}

	if str := tmpl.String(); str != "Hello! {{.Text}} Best regards," {
		t.Errorf("Format: excepted %q but got %q", "Hello! {{.Text}} Best regards,", str)
	}

	if str, err := tmpl.Render(SingleMail{Text: "This is a test."}); err != nil {
		t.Errorf("Execute: failed to render: %s", err.Error())
	} else if str != "Hello! This is a test. Best regards," {
		t.Errorf("Execute: excepted %q but got %q", "Hello! This is a test. Best regards,", str)
	}
}

func TestSingleMail_RenderBody(t *testing.T) {
	var mail SingleMail

	if err := yaml.Unmarshal([]byte(strings.Join([]string{
		"text_template: |",
		"  Hello!",
		"",
		"  {{.Text}}",
		"",
		"  Best regards,",
		"",
		"html_template: |",
		"  <p>Hello!</p>",
		"  {{.Html}}",
		"  <p>Best regards,</p>",
		"",
		"text: This is a test.",
		"html: <p>This is a test.</p>",
	}, "\n")), &mail); err != nil {
		t.Fatalf("Unmarshal: failed to parse: %s", err.Error())
	}

	expected := strings.Join([]string{
		"<p>Hello!</p>",
		"<p>This is a test.</p>",
		"<p>Best regards,</p>",
		"",
		"---------------",
		"Hello!",
		"",
		"This is a test.",
		"",
		"Best regards,",
		"",
	}, "\n")

	if str, err := mail.RenderBody(); err != nil {
		t.Errorf("RenderBody: failed to render: %s", err.Error())
	} else if str != expected {
		t.Errorf("RenderBody:\nexcepted %q\n but got %q", expected, str)
	}
}
