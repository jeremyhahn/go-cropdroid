package service

import (
	"fmt"
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
}

func NewMailer(app *app.App, smtpConfig config.SmtpConfig) common.Mailer {
	return &GMailer{
		app:       app,
		enabled:   smtpConfig.IsEnabled(),
		host:      smtpConfig.GetHost(),
		port:      smtpConfig.GetPort(),
		username:  smtpConfig.GetUsername(),
		password:  smtpConfig.GetPassword(),
		recipient: smtpConfig.GetRecipient()}
}

func (mailer *GMailer) Send(farmName, subject, message string) error {
	if !mailer.enabled {
		mailer.app.Logger.Warningf("[Gmailer.Send] Disabled!")
		return nil
	}

	mailer.app.Logger.Debugf("[GMailer.Send] subject=[%s] %s, message=%s", farmName, subject, message)

	if mailer.host == "" || mailer.port != 0 || mailer.username == "" || mailer.password == "" || mailer.recipient == "" {
		err := fmt.Errorf("Invalid SMTP configuration")
		mailer.app.Logger.Error(err)
		return err
	}

	msg := "From: " + mailer.username + "\n" +
		"To: " + mailer.recipient + "\n" +
		"Subject: " + farmName + " " + subject + "\n\n" + message

	err := smtp.SendMail(mailer.host+":"+strconv.Itoa(mailer.port),
		smtp.PlainAuth("", mailer.username, mailer.password, mailer.host),
		mailer.username, []string{mailer.recipient}, []byte(msg))

	if err != nil {
		mailer.app.Logger.Errorf("smtp error: %s", err)
		return err
	}

	return nil
}
