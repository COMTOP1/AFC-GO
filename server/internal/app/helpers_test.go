package app_test

import (
	"testing"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func testConfig() app.Config {
	return app.Config{
		DomainName: "localhost",
		Session:    auth.Config{CookieName: "session"},
		Passwords: auth.PasswordConfig{
			Iterations: 1, ScryptWorkFactor: 2, ScryptBlockSize: 1, ScryptParallelismFactor: 1, KeyLength: 32,
		},
	}
}

// bareApp wires every route with no database behind it. Only requests that
// guards reject before reaching a store are safe to send.
func bareApp(t *testing.T) *app.App {
	t.Helper()
	a := app.Build(testConfig(), app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)
	return a
}

// guardedGets are GET routes that must reject anonymous users. Append to this
// list whenever a task adds a guarded GET route.
var guardedGets = []string{
	"/api/v1/auth/me",
	"/api/v1/players",
	"/api/v1/users",
	"/api/v1/users/1",
}
