package whatson_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var ctx = context.Background()

func upcoming() whatson.WhatsOn {
	return whatson.WhatsOn{ID: 1, Title: "Presentation night", Content: null.StringFrom("<p>Clubhouse</p>"),
		FileName: null.StringFrom("whatson/old.jpg"), DateOfEvent: today.AddDate(0, 0, 30)}
}

func finished() whatson.WhatsOn {
	return whatson.WhatsOn{ID: 2, Title: "Summer BBQ", DateOfEvent: today.AddDate(0, 0, -30)}
}

func newService(rows ...whatson.WhatsOn) (*whatson.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	objects.Objects["whatson/old.jpg"] = "OLD"
	return whatson.NewService(store, upload.New(objects)), store, objects
}

func TestListByPeriod(t *testing.T) {
	svc, _, _ := newService(upcoming(), finished())
	for period, want := range map[whatson.Period][]int{
		whatson.PeriodAll:    {2, 1},
		whatson.PeriodFuture: {1},
		whatson.PeriodPast:   {2},
		"":                   {2, 1},
	} {
		got, err := svc.List(ctx, period)
		require.NoError(t, err, period)
		ids := make([]int, 0, len(got))
		for _, e := range got {
			ids = append(ids, e.ID)
		}
		assert.Equal(t, want, ids, "period %q", period)
	}
}

func TestListRejectsUnknownPeriod(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.List(ctx, "someday")
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "period")
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, whatson.CreateInput{Title: "", DateOfEvent: today}, nil)
	se, _ := svcerr.As(err)
	require.NotNil(t, se)
	assert.Contains(t, se.Fields, "title")

	_, err = svc.Create(ctx, whatson.CreateInput{Title: "x"}, nil)
	se, _ = svcerr.As(err)
	require.NotNil(t, se)
	assert.Contains(t, se.Fields, "dateOfEvent")
}

func TestCreateSanitises(t *testing.T) {
	svc, store, _ := newService()
	e, err := svc.Create(ctx, whatson.CreateInput{Title: "Quiz", Content: `<p>a</p><script>b</script>`, DateOfEvent: today}, nil)
	require.NoError(t, err)
	assert.Equal(t, "<p>a</p>", store.row(e.ID).Content.String)
}

func TestUpdateRemoveImageDeletesObject(t *testing.T) {
	svc, store, objects := newService(upcoming())
	_, err := svc.Update(ctx, 1, whatson.UpdateInput{RemoveImage: true}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).FileName.Valid)
	assert.Equal(t, []string{"whatson/old.jpg"}, objects.Deleted, "legacy left this object behind; now it is removed")
}

func TestUpdatePartial(t *testing.T) {
	svc, store, _ := newService(upcoming())
	moved := today.AddDate(0, 0, 40)
	_, err := svc.Update(ctx, 1, whatson.UpdateInput{DateOfEvent: &moved}, nil)
	require.NoError(t, err)
	row := store.row(1)
	assert.True(t, row.DateOfEvent.Equal(moved))
	assert.Equal(t, "Presentation night", row.Title)
	assert.Equal(t, "whatson/old.jpg", row.FileName.String)
}

func TestNext(t *testing.T) {
	svc, _, _ := newService(finished())
	_, ok, err := svc.Next(ctx)
	require.NoError(t, err)
	assert.False(t, ok)

	svc, _, _ = newService(upcoming(), finished())
	e, ok, err := svc.Next(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 1, e.ID)
}

func TestDeleteRemovesImage(t *testing.T) {
	svc, _, objects := newService(upcoming())
	e, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Presentation night", e.Title)
	assert.Equal(t, []string{"whatson/old.jpg"}, objects.Deleted)
}
