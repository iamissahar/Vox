package hub

import (
	"context"
	"errors"
	mod "vox/pkg/models"

	"go.uber.org/zap"
)

type HubDB interface {
	GetReference(ctx context.Context, log *zap.Logger, userID string) (filename, text string, err error)
}

type PostgresHub struct{ *mod.Pool }

func NewHubDB(pool *mod.Pool) HubDB {
	return &PostgresHub{Pool: pool}
}

func (ph *PostgresHub) GetReference(ctx context.Context, log *zap.Logger, userID string) (filename, text string, err error) {
	log.Debug("PostgresHub.GetReference", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", ph.Pool == nil), zap.String("userID", userID))
	if ph.Pool == nil || ctx == nil {
		log.Error("Invalid input")
		return filename, text, errors.New("invalid input")
	}

	err = ph.QueryRow(ctx, "SELECT filename, text FROM user_voice WHERE user_id = $1", userID).Scan(&filename, &text)
	if err != nil {
		log.Error("Failed to select from user_voice", zap.Error(err))
		return
	}

	log.Debug("Voice reference retrieved", zap.String("userID", userID), zap.String("filename", filename), zap.String("text", text))
	return
}
