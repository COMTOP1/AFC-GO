// Package emails renders and sends the site's transactional emails.
package emails

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"

	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
)

//go:embed signupEmail.tmpl resetEmail.tmpl
var tmpls embed.FS

// ErrNoMailer means no SMTP server could be reached; callers fall back to
// showing the admin what would have been emailed.
var ErrNoMailer = errors.New("no mailer available")

// Sender sends one email.
type Sender interface {
	Send(ctx context.Context, m mail.Mail) error
}

// SMTP sends via the configured SMTP server, connecting per message.
type SMTP struct {
	init *mail.MailerInit
}

func NewSMTP(init *mail.MailerInit) *SMTP {
	return &SMTP{init: init}
}

func (s *SMTP) Send(ctx context.Context, m mail.Mail) error {
	mailer := s.init.ConnectMailer(ctx)
	if mailer == nil {
		return ErrNoMailer
	}
	defer func() { _ = mailer.Close() }()
	return mailer.SendMail(ctx, m)
}

func parse(name string) (*template.Template, error) {
	t, err := template.New(name).ParseFS(tmpls, name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", name, err)
	}
	return t, nil
}

// Signup is the welcome email carrying a new user's temporary password.
func Signup(to, name, password, domain string) (mail.Mail, error) {
	tpl, err := parse("signupEmail.tmpl")
	if err != nil {
		return mail.Mail{}, err
	}
	return mail.Mail{
		Subject: "Welcome to AFC Aldermaston!",
		Tpl:     tpl,
		To:      to,
		From:    "Aldermaston AFC No-Reply <no-reply.afc@bswdi.co.uk>",
		TplData: struct {
			Name, Email, Password, Domain string
		}{Name: name, Email: to, Password: password, Domain: domain},
	}, nil
}

// Reset is the password reset email.
func Reset(to, resetURL string) (mail.Mail, error) {
	tpl, err := parse("resetEmail.tmpl")
	if err != nil {
		return mail.Mail{}, err
	}
	return mail.Mail{
		Subject: "AFC Security - Reset Password",
		Tpl:     tpl,
		To:      to,
		From:    "AFC Security <no-reply.afc@bswdi.co.uk>",
		TplData: struct {
			Email, URL string
		}{Email: to, URL: resetURL},
	}, nil
}
