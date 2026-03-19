package helpers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func IsValString(ctx *gin.Context, key string) (str string, ok bool) {
	val, _ok := ctx.Get(key)
	if !_ok {
		return str, ok
	}

	switch v := val.(type) {
	case string:
		str = v
		ok = true
	default:
		return str, ok
	}

	return str, ok
}

func CommitOrRollback(ctx context.Context, tx pgx.Tx, err error, log *zap.Logger) {
	log.Debug("CommitOrRollback", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("tx_is_nil", tx == nil), zap.Error(err))
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
	} else {
		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			log.Error("Failed to commit transaction", zap.Error(commitErr))
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}
}
