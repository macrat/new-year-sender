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
	test    = app.Flag("test", "test source file.").Bool()
	dryrun  = app.Flag("dryrun", "test source file and display parse results.").Bool()
)

func verify(source Source) (errored bool) {
	errors := source.Verify()
	for _, e := range errors {
		logrus.Error(e.Error())
	}
	return len(errors) != 0
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

	if *dryrun {
		verify(data)

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
	} else if *test {
		verify(data)
	} else {
		if verify(data) {
			os.Exit(1)
		}

		mailer := NewMailer(data.APIKey)
		mailer.SendAll(data.ToSlice())
	}
}
