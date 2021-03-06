package main

import (
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"
)

type Address mail.Address

func (addr Address) String() string {
	return (*mail.Address)(&addr).String()
}

func (addr Address) IsEmpty() bool {
	return addr.Address == ""
}

func (addr Address) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr *Address) UnmarshalText(data []byte) error {
	if parsed, err := mail.ParseAddress(string(data)); err != nil {
		return err
	} else {
		(*addr).Name = parsed.Name
		(*addr).Address = parsed.Address
		return nil
	}
}

type AddressList []Address

func (list AddressList) StringList() (stringList []string) {
	for _, a := range list {
		stringList = append(stringList, a.String())
	}
	return
}

func (list AddressList) String() string {
	return strings.Join(list.StringList(), ", ")
}

type DateTime struct {
	time.Time
}

func NewDateTime() *DateTime {
	return &DateTime{time.Now()}
}

func (datetime DateTime) RFC822() string {
	return datetime.Format(time.RFC822)
}

func (datetime DateTime) MarshalText() ([]byte, error) {
	return []byte(datetime.Format("2006-01-02 15:04")), nil
}

func (datetime *DateTime) UnmarshalText(data []byte) (err error) {
	tmp, err := time.ParseInLocation("2006-01-02 15:04", string(data), time.Local)
	*datetime = DateTime{tmp}
	return
}

type SingleMail struct {
	Title  string      `yaml:"title,omitempty"`
	Date   *DateTime   `yaml:"date,omitempty"`
	Attach []string    `yaml:"attach,omitempty,flow"`
	Text   string      `yaml:"text,omitempty"`
	Html   string      `yaml:"html,omitempty"`
	From   Address     `yaml:"from,omitempty"`
	To     AddressList `yaml:"to,omitempty,flow"`
	Cc     AddressList `yaml:"cc,omitempty,flow"`
	Bcc    AddressList `yaml:"bcc,omitempty,flow"`
}

func (target SingleMail) Override(source SingleMail) SingleMail {
	if source.Title == "" {
		source.Title = target.Title
	}

	if source.Date == nil {
		source.Date = target.Date
	}

	source.Attach = append(source.Attach, target.Attach...)

	if source.Text == "" {
		source.Text = target.Text
	}

	if source.Html == "" {
		source.Html = target.Html
	}

	if source.From.IsEmpty() {
		source.From = target.From
	}

	source.To = append(source.To, target.To...)
	source.Cc = append(source.Cc, target.Cc...)
	source.Bcc = append(source.Bcc, target.Bcc...)

	return source
}

func (mail SingleMail) BodyString() string {
	switch {
	case len(mail.Text) > 0 && len(mail.Html) == 0:
		return mail.Text
	case len(mail.Text) == 0 && len(mail.Html) > 0:
		return mail.Html
	case len(mail.Text) > 0 && len(mail.Html) > 0:
		return fmt.Sprintf("%s\n---------------\n%s", mail.Html, mail.Text)
	default:
		return ""
	}
}

func (mail SingleMail) String() string {
	return fmt.Sprintf(
		"[%s] %v | from: %s | to: %v | cc: %v | bcc: %v\nattached: %v\n%s\n",
		mail.Title,
		mail.Date,
		mail.From,
		mail.To,
		mail.Cc,
		mail.Bcc,
		mail.Attach,
		mail.BodyString(),
	)
}

type SourceMails struct {
	SingleMail `yaml:",inline"`
	Mails      []SourceMails `yaml:"mails,omitempty"`
}

func (mails SourceMails) Walk(base *SingleMail, fun func(SingleMail)) {
	if base == nil {
		base = &mails.SingleMail
	} else {
		overrided := (*base).Override(mails.SingleMail)
		base = &overrided
	}

	if len(mails.Mails) > 0 {
		for _, m := range mails.Mails {
			m.Walk(base, fun)
		}
	} else {
		fun(*base)
	}
}

func (mails SourceMails) ToSlice() (result []SingleMail) {
	mails.Walk(nil, func(sm SingleMail) {
		result = append(result, sm)
	})
	return
}

type Source struct {
	SourceMails `yaml:",inline"`

	APIKey string `yaml:"apikey"`
}

func (s Source) VerifyAttach() (errors []error) {
	notfounds := make(map[string]struct{})
	s.Walk(nil, func(mail SingleMail) {
		for _, fname := range mail.Attach {
			if f, err := os.Stat(fname); err != nil || !f.Mode().IsRegular() {
				notfounds[fname] = struct{}{}
			}
		}
	})
	for nf, _ := range notfounds {
		errors = append(errors, fmt.Errorf("file notfound: %s", nf))
	}
	return
}

func (s Source) VerifyBody() (errors []error) {
	s.Walk(nil, func(mail SingleMail) {
		if len(mail.Text) == 0 && len(mail.Html) == 0 {
			errors = append(errors, fmt.Errorf("the text and html of the email that to %s is empty; please set least one of text or html", mail.To))
		}
	})
	return
}

func (s Source) Verify() (errors []error) {
	errors = append(errors, s.VerifyAttach()...)
	errors = append(errors, s.VerifyBody()...)
	return
}
