package main

import (
	"fmt"
	"time"
)

type SourceFrom struct {
	Name    string `yaml:"name,omitempty"`
	Address string `yaml:"address,omitempty"`
}

func (from SourceFrom) String() string {
	return fmt.Sprintf("%s <%s>", from.Name, from.Address)
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
	Title  string     `yaml:"title,omitempty"`
	Date   *DateTime  `yaml:"date,omitempty"`
	Attach []string   `yaml:"attach,omitempty,flow"`
	Text   string     `yaml:"text,omitempty"`
	From   SourceFrom `yaml:"from,omitempty"`
	To     []string   `yaml:"to,omitempty,flow"`
	Cc     []string   `yaml:"cc,omitempty,flow"`
	Bcc    []string   `yaml:"bcc,omitempty,flow"`
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

	if source.From.Name == "" {
		source.From.Name = target.From.Name
	}

	if source.From.Address == "" {
		source.From.Address = target.From.Address
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
