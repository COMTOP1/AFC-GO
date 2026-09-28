package setting_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/setting"
)

type fakeStore struct {
	mu   sync.Mutex
	rows map[string]string
}

func newFakeStore(kv map[string]string) *fakeStore {
	if kv == nil {
		kv = map[string]string{}
	}
	return &fakeStore{rows: kv}
}

func (f *fakeStore) GetSetting(_ context.Context, id string) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.rows[id]
	if !ok {
		return setting.Setting{}, fmt.Errorf("failed to get setting: %w", sql.ErrNoRows)
	}
	return setting.Setting{ID: id, SettingText: v}, nil
}

func (f *fakeStore) AddSetting(_ context.Context, s setting.Setting) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[s.ID]; ok {
		return setting.Setting{}, fmt.Errorf("duplicate key %s", s.ID)
	}
	f.rows[s.ID] = s.SettingText
	return s, nil
}

func (f *fakeStore) EditSetting(_ context.Context, s setting.Setting) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows[s.ID] = s.SettingText
	return s, nil
}

func (f *fakeStore) DeleteSetting(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, id)
	return nil
}
