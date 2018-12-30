package main

import (
	"fmt"
	"net/mail"
	"os"
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

type DateTime struct {
	time.Time
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
	Title  string    `yaml:"title,omitempty"`
	Date   *DateTime `yaml:"date,omitempty"`
	Attach []string  `yaml:"attach,omitempty,flow"`
	Text   string    `yaml:"text,omitempty"`
	From   Address   `yaml:"from,omitempty"`
	To     []Address `yaml:"to,omitempty,flow"`
	Cc     []Address `yaml:"cc,omitempty,flow"`
	Bcc    []Address `yaml:"bcc,omitempty,flow"`
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

	if source.From.IsEmpty() {
		source.From = target.From
	}

	source.To = append(source.To, target.To...)
	source.Cc = append(source.Cc, target.Cc...)
	source.Bcc = append(source.Bcc, target.Bcc...)

	return source
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
		mail.Text,
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

func (s Source) VerifyBody() error {
	count := 0
	s.Walk(nil, func(mail SingleMail) {
		if len(mail.Text) == 0 {
			count += 1
		}
	})
	if count > 0 {
		return fmt.Errorf("text can't be empty: there is %d empty mails.", count)
	}
	return nil
}

func (s Source) Verify() (errors []error) {
	errors = append(errors, s.VerifyAttach()...)

	if err := s.VerifyBody(); err != nil {
		errors = append(errors, err)
	}

	return
}
