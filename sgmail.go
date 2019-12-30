package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

type Mailer struct {
	client *sendgrid.Client
}

func NewMailer(apiKey string) Mailer {
	return Mailer{client: sendgrid.NewSendClient(apiKey)}
}

func (mailer Mailer) sendSGV3(m *mail.SGMailV3) error {
	result, err := mailer.client.Send(m)
	if err != nil {
		return err
	} else if result.StatusCode != 202 {
		return fmt.Errorf("unknown error: %d: %#v", result.StatusCode, result.Body)
	} else {
		return nil
	}
}

func (mailer Mailer) Send(m SingleMail) error {
	return mailer.sendSGV3(SingleMailToSendGrid(m))
}

func (mailer Mailer) SendAll(ms []SingleMail) {
	mails := make([]*mail.SGMailV3, len(ms))
	for i, m := range ms {
		mails[i] = SingleMailToSendGrid(m)
	}

	for len(mails) > 0 {
		m := mails[0]
		mails = mails[1:]

		if err := mailer.sendSGV3(m); err != nil {
			mails = append(mails, m)
			logrus.Errorf("failed to send: %s", err.Error())
		}
	}
}

func ReadAttachment(filename string) *mail.Attachment {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Fatalf("failed open attachment: %s", err.Error())
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Fatalf("failed read attachment: %s", err.Error())
	}

	mimeType := http.DetectContentType(data)
	encoded := base64.StdEncoding.EncodeToString(data)

	attach := mail.NewAttachment()
	attach.SetContent(encoded)
	attach.SetType(mimeType)
	attach.SetFilename(path.Base(filename))
	attach.SetDisposition("inline")
	attach.SetContentID(path.Base(filename))

	return attach
}

func SingleMailToSendGrid(source SingleMail) *mail.SGMailV3 {
	result := mail.NewV3Mail()

	result.Subject = source.Title

	if source.Date == nil {
		source.Date = NewDateTime()
	}
	result.SetHeader("Date", source.Date.RFC822())
	result.SetSendAt(int(source.Date.Unix()))

	for _, fname := range source.Attach {
		result.AddAttachment(ReadAttachment(fname))
	}

	result.AddContent(mail.NewContent("text/plain", source.Text))

	result.SetFrom(mail.NewEmail(source.From.Name, source.From.Address))

	ps := mail.NewPersonalization()
	for _, t := range source.To {
		ps.AddTos(mail.NewEmail(t.Name, t.Address))
	}
	for _, c := range source.Cc {
		ps.AddCCs(mail.NewEmail(c.Name, c.Address))
	}
	for _, b := range source.Bcc {
		ps.AddBCCs(mail.NewEmail(b.Name, b.Address))
	}
	result.AddPersonalizations(ps)

	return result
}
