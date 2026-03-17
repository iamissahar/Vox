package voice

import (
	"context"
	"errors"
	mod "vox/pkg/models"

	"go.uber.org/zap"
)

type VoiceDB interface {
	SaveNewVoiceReference(ctx context.Context, log *zap.Logger, userID, text, fileID, path, typeof string) error
}

type PostgresVoice struct{ *mod.Pool }

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
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO files (id, full_path, type, text) VALUES ($1, $2, $3, $4)", fileID, path, typeof, text)
	if err != nil {
		log.Error("Failed to insert files", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO files_and_users (file_id, user_id, is_active) VALUES ($1, $2, $3)", fileID, userID, true)
	if err != nil {
		log.Error("Failed to insert files_and_users", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	log.Debug("New voice reference saved")
	return err
}
