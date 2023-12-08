package image

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getImages(ctx context.Context) ([]Image, error) {
	var imagesDB []Image
	builder := sq.Select("id", "file_name", "caption").
		From("afc.images").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getImages: %w", err))
	}
	err = s.db.SelectContext(ctx, &imagesDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	return imagesDB, nil
}

func (s *Store) getImage(ctx context.Context, imageParam Image) (Image, error) {
	var imageDB Image
	builder := utils.PSQL().Select("id", "file_name", "caption").
		From("afc.images").
		Where(sq.Eq{"id": imageParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getImage: %w", err))
	}
	err = s.db.GetContext(ctx, &imageDB, sql, args...)
	if err != nil {
		return Image{}, fmt.Errorf("failed to get image: %w", err)
	}
	return imageDB, nil
}

func (s *Store) addImage(ctx context.Context, imageParam Image) (Image, error) {
	builder := utils.PSQL().Insert("afc.images").
		Columns("file_name", "caption").
		Values(imageParam.FileName, imageParam.Caption)
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
	imageParam.ID = int(id)
	return imageParam, nil
}

func (s *Store) deleteImage(ctx context.Context, imageParam Image) error {
	builder := utils.PSQL().Delete("afc.images").
		Where(sq.Eq{"id": imageParam.ID})
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
