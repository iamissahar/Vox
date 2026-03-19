package voice

import (
	"context"
	"errors"
	"vox/pkg/helpers"
	mod "vox/pkg/models"

	"go.uber.org/zap"
)

type VoiceDB interface {
	SaveNewVoiceReference(ctx context.Context, log *zap.Logger, userID, text, fileID, path, typeof string) error
	GetVoiceReference(ctx context.Context, log *zap.Logger, userID string) (arr [5]VoiceReference, n int, err error)
	DeleteVoiceReference(ctx context.Context, log *zap.Logger, userID, fileID string) error
}

type PostgresVoice struct{ *mod.Pool }

type VoiceReference struct {
	FileID string `json:"file_id"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	Text   string `json:"text"`
}

func NewVoiceDB(pool *mod.Pool) VoiceDB {
	return &PostgresVoice{Pool: pool}
}

func (v *PostgresVoice) SaveNewVoiceReference(ctx context.Context, log *zap.Logger, userID, text, fileID, path, typeof string) (err error) {
	log.Debug("PostgresVoice.SaveNewVoiceReference",
		zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", v.Pool == nil),
		zap.String("user_id", userID), zap.String("text", text),
		zap.String("file_id", fileID), zap.String("path", path),
		zap.String("typeof", typeof),
	)
	if v.Pool == nil || ctx == nil {
		log.Error("Invalid input")
		return errors.New("invalid input")
	}

	tx, err := v.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return
	}

	defer helpers.CommitOrRollback(ctx, tx, err, log)

	_, err = tx.Exec(ctx, "INSERT INTO files (id, full_path, type, text) VALUES ($1, $2, $3, $4)", fileID, path, typeof, text)
	if err != nil {
		log.Error("Failed to insert files", zap.Error(err))
		return
	}

	_, err = tx.Exec(ctx, "INSERT INTO files_and_users (file_id, user_id, is_active) VALUES ($1, $2, $3)", fileID, userID, true)
	if err != nil {
		log.Error("Failed to insert files_and_users", zap.Error(err))
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", zap.Error(err))
		return
	}

	log.Debug("New voice reference saved")
	return
}

func (v *PostgresVoice) GetVoiceReference(ctx context.Context, log *zap.Logger, userID string) (arr [5]VoiceReference, n int, err error) {
	log.Debug("PostgresVoice.GetVoiceReference",
		zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", v.Pool == nil),
		zap.String("user_id", userID),
	)

	if v.Pool == nil || ctx == nil {
		log.Error("Invalid input")
		return [5]VoiceReference{}, 0, errors.New("invalid input")
	}

	tx, err := v.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return
	}

	defer helpers.CommitOrRollback(ctx, tx, err, log)

	rows, err := tx.Query(ctx, "SELECT file_id FROM files_and_users WHERE user_id = $1", userID)
	if err != nil {
		log.Error("Failed to select from files_and_users", zap.Error(err))
		return
	}

	for i := 0; i < 5 && rows.Next(); i++ {
		err = rows.Scan(&arr[i].FileID)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return
		}
		n = i
	}
	rows.Close()
	err = rows.Err()
	if err != nil {
		log.Error("Failed to close rows", zap.Error(err))
	}

	fileIDs := make([]string, 0, n+1)
	for _, ref := range arr[:n+1] {
		fileIDs = append(fileIDs, ref.FileID)
	}

	rows, err = tx.Query(ctx, "SELECT full_path, type, text FROM files WHERE id = ANY($1)", fileIDs)
	if err != nil {
		log.Error("Failed to select from files", zap.Error(err))
		return
	}

	for i := 0; i < 5 && rows.Next(); i++ {
		err = rows.Scan(&arr[i].Path, &arr[i].Type, &arr[i].Text)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return
		}
	}

	rows.Close()
	err = rows.Err()
	if err != nil {
		log.Error("Failed to close rows", zap.Error(err))
	}
	return
}

func (v *PostgresVoice) DeleteVoiceReference(ctx context.Context, log *zap.Logger, userID, fileID string) (err error) {
	log.Debug("PostgresVoice.DeleteVoiceReference",
		zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", v.Pool == nil),
		zap.String("user_id", userID), zap.String("file_id", fileID),
	)

	if v.Pool == nil || ctx == nil {
		log.Error("Invalid input")
		return errors.New("invalid input")
	}

	tx, err := v.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return
	}

	defer helpers.CommitOrRollback(ctx, tx, err, log)

	_, err = tx.Exec(ctx, "UPDATE files_and_users SET is_active = false WHERE user_id = $1 AND file_id = $2", userID, fileID)
	if err != nil {
		log.Error("Failed to update files_and_users", zap.Error(err))
		return
	}
	return
}
