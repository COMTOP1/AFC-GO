package image

import (
	"context"
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	sq "github.com/Masterminds/squirrel"
)

func (s *Store) getImages(ctx context.Context) ([]Image, error) {
	var i []Image
	builder := sq.Select("id", "image", "file_name", "caption").
		From("afc.images").
		OrderBy("id")
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getImages: %w", err))
	}
	err = s.db.SelectContext(ctx, &i, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	return i, nil
}

func (s *Store) getImage(ctx context.Context, i Image) (Image, error) {
	var i1 Image
	builder := sq.Select("id", "image", "file_name", "caption").
		From("afc.images").
		Where(sq.Eq{"id": i.ID})
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getImage: %w", err))
	}
	err = s.db.SelectContext(ctx, &i1, sql)
	if err != nil {
		return Image{}, fmt.Errorf("failed to get image: %w", err)
	}
	return i1, nil
}

func (s *Store) addImage(ctx context.Context, i Image) (Image, error) {
	builder := utils.MySQL().Insert("afc.images").
		Columns("image", "file_name", "caption").
		Values(i.Image, i.FileName, i.Caption)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addImage: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Image{}, fmt.Errorf("failed to add image: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Image{}, fmt.Errorf("failed to add image: %w", err)
	}
	if rows < 1 {
		return Image{}, fmt.Errorf("failed to add image: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Image{}, fmt.Errorf("failed to add image: %w", err)
	}
	i.ID = int(id)
	return i, nil
}

func (s *Store) deleteImage(ctx context.Context, d Image) error {
	builder := utils.MySQL().Delete("afc.images").
		Where(sq.Eq{"id": d.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteImage: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}
