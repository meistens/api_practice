package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// variable with type embed.FS to hold the email
// templates
// this has a comment directive in the format `go:embed "<path>`
// IMMEDIATELY above it, which means we want to store
// the contents of the .templates dir in the templateFS
// embedded file system (FS) variable

//go:embed "templates"
var templateFS embed.FS

// mailer struct
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	//init. a new mail.dialer instance with smtp server settings, with a 5s timeout
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	//return a mailer instance containing the dialer and sender info
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// define a send() method on mailer type, which takes a recipient
// mail addr as the first param, name of file with templates
// and any dynamic data for the templates as an interface{} param
func (m Mailer) Send(recipient, templateFile string, data any) error {
	// use parseFS() method to parse the required template
	// file from the embedded file system
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// execute named template "subject", passing in the
	// dynamic data and storing the result in a
	// bytes.buffer variable
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// follow same pattern to exec. "plainbody" and store the
	// result in the plainbody variable
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// add likewise with the "htmlBody" template
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then we use the SetHeader() method to set the email recipient, sender and subject
	// headers, the SetBody() method to set the plain-text body, and the AddAlternative()
	// method to set the HTML body. It's important to note that AddAlternative() should
	// always be called *after* SetBody().
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Call the DialAndSend() method on the dialer, passing in the message to send. This
	// opens a connection to the SMTP server, sends the message, then closes the
	// connection. If there is a timeout, it will return a "dial tcp: i/o timeout"
	// error.
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}
