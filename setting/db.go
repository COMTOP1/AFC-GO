package setting

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getSettings(ctx context.Context) ([]Setting, error) {
	var settingsDB []Setting
	builder := sq.Select("id", "setting_text").
		From("afc.settings").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get settings: %w", err))
	}
	err = s.db.SelectContext(ctx, &settingsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	return settingsDB, nil
}

func (s *Store) getSetting(ctx context.Context, settingID string) (Setting, error) {
	var settingDB Setting
	builder := utils.PSQL().Select("id", "setting_text").
		From("afc.settings").
		Where(sq.Eq{"id": settingID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get setting: %w", err))
	}
	err = s.db.GetContext(ctx, &settingDB, sql, args...)
	if err != nil {
		return Setting{}, fmt.Errorf("failed to get setting: %w", err)
	}
	return settingDB, nil
}

func (s *Store) addSetting(ctx context.Context, settingParam Setting) (Setting, error) {
	builder := utils.PSQL().Insert("afc.settings").
		Columns("setting_text").
		Values(settingParam.SettingText)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addSetting: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Setting{}, fmt.Errorf("failed to add setting: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Setting{}, fmt.Errorf("failed to add setting: %w", err)
	}
	return settingParam, nil
}

func (s *Store) editSetting(ctx context.Context, settingParam Setting) (Setting, error) {
	builder := utils.PSQL().Update("afc.settings").
		SetMap(map[string]interface{}{
			"setting_text": settingParam.SettingText,
		}).
		Where(sq.Eq{"id": settingParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editSetting: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Setting{}, fmt.Errorf("failed to edit setting: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Setting{}, fmt.Errorf("failed to edit setting: %w", err)
	}
	return settingParam, nil
}

func (s *Store) deleteSetting(ctx context.Context, settingID string) error {
	builder := utils.PSQL().Delete("afc.settings").
		Where(sq.Eq{"id": settingID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteSetting: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}
	return nil
}
