package setting_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func TestInfo(t *testing.T) {
	store := newFakeStore(nil)
	svc := setting.NewService(store)

	got, err := svc.Info(ctx)
	require.NoError(t, err)
	assert.Empty(t, got, "unset is empty, not an error")

	saved, err := svc.SetInfo(ctx, `<div>History</div><script>x</script>`)
	require.NoError(t, err)
	assert.Equal(t, "<div>History</div>", saved)
	saved, err = svc.SetInfo(ctx, `<div>Updated</div>`) // second write edits rather than inserts
	require.NoError(t, err)
	assert.Equal(t, "<div>Updated</div>", store.rows["infoContent"])
	assert.Equal(t, saved, store.rows["infoContent"])
}

func TestDisplayEmail(t *testing.T) {
	store := newFakeStore(map[string]string{"displayEmail": "old@example.test"})
	svc := setting.NewService(store)

	_, err := svc.SetDisplayEmail(ctx, "not-an-email")
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "email")
	assert.Equal(t, "old@example.test", store.rows["displayEmail"])

	_, err = svc.SetDisplayEmail(ctx, "hello@example.test")
	require.NoError(t, err)
	got, _ := svc.DisplayEmail(ctx)
	assert.Equal(t, "hello@example.test", got)

	_, err = svc.SetDisplayEmail(ctx, "")
	require.NoError(t, err)
	_, present := store.rows["displayEmail"]
	assert.False(t, present, "empty removes the setting, as legacy does")
}

func TestSettingRoutes(t *testing.T) {
	svc := setting.NewService(newFakeStore(map[string]string{"infoContent": "<div>Hi</div>"}))
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	setting.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Treasurer))
	secretary := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))

	rec := anon.Get(t, "/api/v1/info")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<div>Hi</div>", apitest.Decode[setting.InfoContent](t, rec).Content)

	assert.Equal(t, http.StatusUnauthorized, anon.JSON(t, http.MethodPut, "/api/v1/info", setting.InfoContent{Content: "x"}).Code)
	assert.Equal(t, http.StatusOK, editor.JSON(t, http.MethodPut, "/api/v1/info", setting.InfoContent{Content: "<p>x</p>"}).Code)

	assert.Equal(t, http.StatusForbidden, editor.JSON(t, http.MethodPut, "/api/v1/settings/display-email", setting.DisplayEmail{Email: "a@b.test"}).Code,
		"treasurer is an editor but not club secretary or higher")
	rec = secretary.JSON(t, http.MethodPut, "/api/v1/settings/display-email", setting.DisplayEmail{Email: "a@b.test"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "a@b.test", apitest.Decode[setting.DisplayEmail](t, rec).Email)
}
