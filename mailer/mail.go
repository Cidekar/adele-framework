package mailer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/CloudyKit/jet/v6"
	apimail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

// Listen on the mail channel and send when a payload is received.
// The method will run continually in the background and send error
// or success messages on the results channel.
func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs // listen for jobs
		err := m.Send(msg)

		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

// Determines the appropriate method for sending an email message.
// If a third-party API is configured (API name is set and not "smtp", and both APIKey and APIUrl are provided),
// it delegates the sending to chooseAPI. Otherwise, it defaults to sending the message via SMTP.
func (m *Mail) Send(msg Message) error {
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		return m.chooseAPI(msg)
	}
	return m.SendSMTPMessage(msg)
}

// Choose the API method for sending an email message.
// If a third-party supported API is configured, it delegates to the corresponding it.
// Otherwise, it will return an error.
func (m *Mail) chooseAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("unknown api %s; only mailgun, sparkpost, or sendgrid accepted", m.API)
	}
}

// Sends an email using a specified API transport.
// It builds the email message, sets default sender values if missing,
// attaches files, and uses the configured API client to transmit the message.
func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	config := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(transport, config)
	if err != nil {
		return err
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	transmission := &apimail.Transmission{
		Recipients: []string{msg.To}, // TODO support sending to multiple
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	// Add attachments
	err = m.addAPIAttachments(msg, transmission)
	if err != nil {
		return err
	}

	// Send the mail
	_, err = driver.Send(transmission)
	if err != nil {
		return err
	}

	return nil
}

// Read and attache files from the message to the API transmission.
// It loads each file's content and appends it to the transmission's attachments list.
func (m *Mail) addAPIAttachments(msg Message, transmission *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []apimail.Attachment

		for _, x := range msg.Attachments {
			var attach apimail.Attachment
			content, err := ioutil.ReadFile(x)
			if err != nil {
				return err
			}

			fileName := filepath.Base(x)
			attach.Bytes = content
			attach.Filename = fileName
			attachments = append(attachments, attach)
		}

		transmission.Attachments = attachments
	}
	return nil
}

// Sends an email using SMTP with HTML and plain text bodies.
// It sets default sender values, configures the SMTP client, attaches any files,
// and sends the message using the SMTP protocol.
func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

// Get the appropriate encryption type based on a string value.
// It maps a string value to the corresponding mail encryption type.
// Defaults to STARTTLS if the input is unrecognized.
func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

// Generates the HTML body for an email using a Jet template if it exists,
// falling back to a Go template otherwise.
func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	tmply := fmt.Sprintf("%s/%s.html.jet", m.Templates, msg.Template)

	var _, err = os.Stat(tmply)

	if !os.IsExist(err) {
		formattedMessage, err := m.buildJetEmail(msg)
		if err != nil {
			return "", err
		}
		return formattedMessage, nil
	}

	formattedMessage, err := m.buildGoEmail(msg)
	if err != nil {
		return "", err
	}
	return formattedMessage, nil

}

// Generates the HTML body for an email using a go template.
func (m *Mail) buildGoEmail(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template) // Go templates

	log.Println("templateToRender: " + templateToRender)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

// Render an HTML email using a Jet template engine.
// It loads the specified template, injects message data, and returns the rendered string.
func (m *Mail) buildJetEmail(msg Message) (string, error) {

	vars := make(jet.VarMap)
	vars.Set("data", msg.Data)

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/", m.Templates)),
	)
	t, err := views.GetTemplate(fmt.Sprintf("%s.html.jet", msg.Template))

	if err != nil {
		return "", err
	}

	var w bytes.Buffer

	if err = t.Execute(&w, vars, nil); err != nil {
		return "", err
	}

	return w.String(), nil
}

// Generates the plain text body for an email using a Jet template if available,
// falling back to a Go template if the Jet template doesn't exist.
func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	tmply := fmt.Sprintf("%s/%s.plain.jet", m.Templates, msg.Template)

	var _, err = os.Stat(tmply)

	if !os.IsExist(err) {
		formattedMessage, err := m.buildJetPlainTextMessage(msg)
		if err != nil {
			return "", err
		}
		return formattedMessage, nil
	}

	formattedMessage, err := m.buildGoPlainTextMessage(msg)
	if err != nil {
		return "", err
	}
	return formattedMessage, nil
}

// Render the plain text body of an email using a Go template.
// It loads the specified .plain.tmpl file and injects the message data into the "body" template block.
func (m *Mail) buildGoPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

// Render the plain text body of an email using a Jet template.
// It loads the specified .plain.jet file, injects the message data, and returns the rendered output.
func (m *Mail) buildJetPlainTextMessage(msg Message) (string, error) {
	vars := make(jet.VarMap)
	vars.Set("data", msg.Data)

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/", m.Templates)),
	)
	t, err := views.GetTemplate(fmt.Sprintf("%s.plain.jet", msg.Template))

	if err != nil {
		return "", err
	}

	var w bytes.Buffer

	if err = t.Execute(&w, vars, nil); err != nil {
		return "", err
	}

	return w.String(), nil
}

// Process the given HTML string and inlines its CSS using Premailer.
// Returns the transformed HTML with styles applied inline.
func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
