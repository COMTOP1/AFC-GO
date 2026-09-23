package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/infrastructure/mail")

type (
	// Mailer is the struct used to send mail, can only be used once connected to mailer
	Mailer struct {
		*mail.SMTPClient
		Defaults Defaults
	}

	// MailerInit is the config store for the mailer, can be initialised once then connected multiple times
	MailerInit struct {
		SMTPServer mail.SMTPServer
		Defaults   Defaults
	}

	// Defaults is the default values for the mailer if none are explicitly mentioned
	Defaults struct {
		DefaultTo   string
		DefaultCC   []string
		DefaultBCC  []string
		DefaultFrom string
	}

	// Config represents a configuration to connect to an SMTP server
	Config struct {
		Host     string
		Port     int
		Username string
		Password string
		Defaults Defaults
	}

	// Mail represents an email to be sent
	Mail struct {
		Subject string
		To      string
		Cc      []string
		Bcc     []string
		From    string
		Error   error
		Tpl     *template.Template
		TplData interface{}
	}
)

// NewMailer creates a new SMTP client
func NewMailer(config Config) *MailerInit {
	smtpServer := mail.SMTPServer{
		Host:           config.Host,
		Port:           config.Port,
		Username:       config.Username,
		Password:       config.Password,
		Encryption:     mail.EncryptionSTARTTLS,
		Authentication: mail.AuthLogin,
		ConnectTimeout: 10 * time.Second,
		SendTimeout:    10 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: config.Host,
		},
	}

	return &MailerInit{
		SMTPServer: smtpServer,
		Defaults:   config.Defaults,
	}
}

// ConnectMailer connects to the mail server
func (m *MailerInit) ConnectMailer(ctx context.Context) *Mailer {
	_, span := tracer.Start(ctx, "mail.ConnectMailer")
	defer span.End()
	smtpClient, err := m.SMTPServer.Connect()
	if err != nil {
		span.RecordError(err)
		slog.Info(fmt.Sprintf("mailer failed: %+v", err))
		return nil
	}
	slog.Info("connected to mailer: " + m.SMTPServer.Host)
	return &Mailer{smtpClient, m.Defaults}
}

// CheckSendable verifies that the email can be sent
func (m *Mailer) CheckSendable(item Mail) error {
	if item.To == "" {
		return errors.New("no To field is set")
	}
	return nil
}

// SendMail sends a template email
func (m *Mailer) SendMail(ctx context.Context, item Mail) error {
	_, span := tracer.Start(ctx, "mail.SendMail")
	defer span.End()

	err := m.CheckSendable(item)
	if err != nil {
		span.RecordError(err)
		return err
	}
	to, from, cc, bcc := m.setEmailHeader(item)
	body := bytes.Buffer{}
	err = item.Tpl.Execute(&body, item.TplData)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to exec tpl: %w", err)
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(item.Subject)
	if len(item.Cc) != 0 {
		email.AddCc(cc...)
	}
	if len(item.Bcc) != 0 {
		email.AddBcc(bcc...)
	}
	email.SetBody(mail.TextHTML, body.String())
	if email.Error != nil {
		span.RecordError(email.Error)
		return fmt.Errorf("failed to set mail data: %w", email.Error)
	}
	err = email.Send(m.SMTPClient)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (m *Mailer) setEmailHeader(item Mail) (to, from string, cc, bcc []string) {
	if len(item.To) > 0 {
		to = item.To
	} else {
		to = m.Defaults.DefaultTo
	}

	if len(item.Cc) > 0 {
		cc = item.Cc
	} else {
		cc = m.Defaults.DefaultCC
	}

	if len(item.Bcc) > 0 {
		bcc = item.Bcc
	} else {
		bcc = m.Defaults.DefaultBCC
	}

	if len(item.From) > 0 {
		from = item.From
	} else {
		from = m.Defaults.DefaultFrom
	}
	return
}
