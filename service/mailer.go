package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strconv"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
)

type GMailer struct {
	app       *app.App
	enabled   bool
	host      string
	port      int
	username  string
	password  string
	recipient string
	common.Mailer
}

func NewMailer(app *app.App) common.Mailer {
	return CreateMailer(app, app.Smtp)
}

func CreateMailer(app *app.App, smtpConfig *config.Smtp) common.Mailer {
	if smtpConfig == nil {
		if app.Smtp != nil {
			smtpConfig = app.Smtp
		} else {
			smtpConfig = &config.Smtp{Enable: false}
		}
	}
	return &GMailer{
		app:       app,
		enabled:   smtpConfig.IsEnabled(),
		host:      smtpConfig.GetHost(),
		port:      smtpConfig.GetPort(),
		username:  smtpConfig.GetUsername(),
		password:  smtpConfig.GetPassword(),
		recipient: smtpConfig.GetRecipient()}
}

func (mailer *GMailer) SetRecipient(recipient string) {
	mailer.recipient = recipient
}

func (mailer *GMailer) Send(subject, message string) error {
	if !mailer.enabled {
		mailer.app.Logger.Warningf("Disabled!")
		return nil
	}

	mailer.app.Logger.Debugf("subject=[%s] %s, message=%s", subject, message)

	if mailer.host == "" || mailer.port <= 0 || mailer.username == "" || mailer.password == "" || mailer.recipient == "" {
		err := fmt.Errorf("invalid SMTP configuration")
		mailer.app.Logger.Error(err)
		return err
	}

	msg := "From: " + mailer.username + "\n" +
		"To: " + mailer.recipient + "\n" +
		"Subject: " + subject + "\n\n" + message

	err := smtp.SendMail(mailer.host+":"+strconv.Itoa(mailer.port),
		smtp.PlainAuth("", mailer.username, mailer.password, mailer.host),
		mailer.username, []string{mailer.recipient}, []byte(msg))

	if err != nil {
		mailer.app.Logger.Errorf("smtp error: %s", err)
		return err
	}

	return nil
}

func (mailer *GMailer) SendHtml(template, subject string, data interface{}) (bool, error) {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body, err := mailer.parseTemplate(template, data)
	if err != nil {
		return false, err
	}

	// mailer.app.Logger.Debugf("body=%s", body)
	// mailer.app.Logger.Debugf("data=%+v", data)

	msg := "From: " + mailer.username + "\n" +
		"To: " + mailer.recipient + "\n" +
		"Subject: " + subject + "\n" + mime + "\n" + body

	if err := smtp.SendMail(mailer.host+":"+strconv.Itoa(mailer.port),
		smtp.PlainAuth("", mailer.username, mailer.password, mailer.host),
		mailer.username, []string{mailer.recipient}, []byte(msg)); err != nil {

		return false, err
	}

	return true, nil
}

func (mailer *GMailer) parseTemplate(templateFileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
