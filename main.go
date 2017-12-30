package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	app     = kingpin.New("new-year-sender", "The new year email sender.")
	source  = app.Flag("source", "source yaml file.").File()
	verbose = app.Flag("verbose", "verbose output for debug.").Bool()
	test    = app.Flag("test", "test source file").Bool()
)

func fileCheck(mails Source) (notfounds map[string]struct{}) {
	notfounds = make(map[string]struct{})
	mails.Walk(nil, func(mail SingleMail) {
		for _, fname := range mail.Attach {
			if f, err := os.Stat(fname); err != nil || !f.Mode().IsRegular() {
				notfounds[fname] = struct{}{}
			}
		}
	})
	return
}

func textCheck(mails Source) error {
	count := 0
	mails.Walk(nil, func(mail SingleMail) {
		if len(mail.Text) == 0 {
			count += 1
		}
	})
	if count > 0 {
		return fmt.Errorf("text can't be empty: there is %d empty mails.", count)
	}
	return nil
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *source == nil {
		*source = os.Stdin
	}

	if *verbose {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	logrus.Infof("read source: ", (*source).Name())

	raw, err := ioutil.ReadAll(*source)
	if err != nil {
		logrus.Fatalf("failed to read source: %s: %s", (*source).Name(), err.Error())
		return
	}

	data := Source{}
	if err = yaml.Unmarshal(raw, &data); err != nil {
		logrus.Fatalf("failed to parse source: %s: %s", (*source).Name(), err.Error())
		return
	}

	data.Walk(nil, func(mail SingleMail) {
		logrus.Info(mail)
	})

	if *test {
		for x, _ := range fileCheck(data) {
			logrus.Warnf("file notfound: %s", x)
		}
		if err = textCheck(data); err != nil {
			logrus.Error(err.Error())
		}

		for i, mail := range data.ToSlice() {
			if i != 0 {
				fmt.Println(strings.Repeat("=", 30))
			}

			fmt.Println("title: ", mail.Title)
			fmt.Println("from: ", mail.From)
			fmt.Printf("to: %v\n", mail.To)
			fmt.Printf("cc: %v\n", mail.Cc)
			fmt.Printf("bcc: %v\n", mail.Bcc)
			fmt.Println("date: ", mail.Date)
			fmt.Printf("Attached: %v\n", strings.Join(mail.Attach, ", "))
			fmt.Println()
			fmt.Println(mail.Text)
		}
	} else {
		if err = textCheck(data); err != nil {
			logrus.Fatal(err.Error())
		}

		mailer := NewMailer(data.APIKey)
		mailer.SendAll(data.ToSlice())
	}
}
