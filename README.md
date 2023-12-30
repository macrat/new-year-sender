new-year-sender
===============

The email sender for new year email.

## Simple usage
Install new-year-sender,

``` shell
$ go get github.com/macrat/new-year-sender
```

Make source file like this,

``` yaml
apikey: your-API-key-of-SendGrid

from: your name <your-email@example.com>

to:
  - name <destination-email@example.com>

date: 2018-01-01 00:00  # the date to send email.

title: E-Mail subject

text: |
  contents of an email.

  this is test

html: |
  <p>contents of an <b>email</b>.</p>

  <p>this is test</p>

attach:
  - path/to/file.ext
```

And send it.

``` shell
$ new-year-sender --source source-file.yml
```

## Send many emails
You can send many emails with the simple source file.
This behavior is like an object-based programming.

for example:

``` yaml
apikey: your-API-key-of-SendGrid

# common settings
from: your name <your-email@example.com>

date: 2018-01-01 00:00

title: hello

text: |
  hello!
  this is test e-mail!!

attach:
  - attached-file.png

# personal settings
mails:
  - title: hello alice  # override title
    to:
      - alice <alice@example.com>

  - attach:   # append attach
      - attached-file2.png

    to:
      - bob@example.com

  # more extend
  - date: 2018-01-01 10:00

    mails:
      - to:
          - charie@example.com
          - dave@example.com
        cc:
          - charie2@example.com

      - to:
        - eve <eve@example.com>
```

This source will send 4 emails.

If you want test source, please use `--test` or `--dryrun` option.
`--test` option will validate source and display errors if that exists.
`--dryrun` option will execute `--test` and display parsed results.
Won't send emails if enabled either option.

Print like this when given the above source file with `--dryrun` option.

``` shell
$ new-year-sender --test < test.yml
title:  hello alice
from:  your name <your-email@example.com>
to: ["alice" <alice@example.com>]
cc: []
bcc: []
date:  2018-01-01 00:00:00 +0900 JST
Attached: attached-file.png

hello!
this is test e-mail!!

==============================
title:  hello
from:  your name <your-email@example.com>
to: [<bob@example.com>]
cc: []
bcc: []
date:  2018-01-01 00:00:00 +0900 JST
Attached: attached-file2.png, attached-file.png

hello!
this is test e-mail!!

==============================
title:  hello
from:  your name <your-email@example.com>
to: [<charie@example.com>, <dave@example.com>]
cc: [<charie2@example.com>]
bcc: []
date:  2018-01-01 10:00:00 +0900 JST
Attached: attached-file.png

hello!
this is test e-mail!!

==============================
title:  hello
from:  your name <your-email@example.com>
to: ["eve" <eve@example.com>]
cc: []
bcc: []
date:  2018-01-01 10:00:00 +0900 JST
Attached: attached-file.png

hello!
this is test e-mail!!
```

## Apply templates
Sender supports template for text and HTML body.

for example:

``` yaml
apikey: your-API-key-of-SendGrid

title: Template example

from: your name <your-email@example.com>

text_template: |
  Hello!

  {{.Text}}

  Best regards,

html_template: |
  <body style="background-color: #f0f0f0; color: #000;">
    <p>Hello!</p>
    {{.Html}}
    <p>Best regards,</p>
  </body>

mails:
  - to: [alice <alice@example.com>]
    text:
      How are you alice?
    html:
      <p>How are you <b>alice</b>?</p>

  - to: [bob <bob@example.com>]
    text:
      booooooooob
    html:
      <p style="font-size:500%">booooooob</p>
```
