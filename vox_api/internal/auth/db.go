package auth

import (
	"context"
	"errors"
	mod "vox/pkg/models"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AuthDB interface {
	AddNewManualUser(ctx context.Context, log *zap.Logger, u UserInfo, hash []byte) (err error)
	GetUser(ctx context.Context, log *zap.Logger, providerID int, userProviderID string) (u UserInfo, err error)
	AddNewProviderUser(ctx context.Context, log *zap.Logger, u UserInfo) (err error)
	GetPasswordHash(ctx context.Context, log *zap.Logger, login string) (hash []byte, err error)
	SaveRefreshToken(ctx context.Context, log *zap.Logger, login, refreshHash string) (err error)
	GetRefreshToken(ctx context.Context, log *zap.Logger, login string) (refreshHash string, err error)
}

type PostgresAuth struct{ *mod.Pool }

func NewAuthDB(pool *mod.Pool) AuthDB {
	return &PostgresAuth{Pool: pool}
}

func (pa *PostgresAuth) AddNewManualUser(ctx context.Context, log *zap.Logger, u UserInfo, hash []byte) (err error) {
	log.Debug("PostgresAuth.AddNewManualUser", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", pa.Pool == nil), zap.Any("u", u), zap.Int("hash_length", len(hash)))
	if pa.Pool == nil || ctx == nil || u.ID == "" || hash == nil {
		log.Error("Invalid input")
		return errors.New("invalid input")
	}

	tx, err := pa.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO users (id, email, name, picture_url) VALUES ($1, $2, $3, $4)", u.ID, u.Email, u.Name, u.Picture)
	if err != nil {
		log.Error("Failed to insert into users", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO auth_references (user_id, password_hash) VALUES ($1, $2)", u.ID, hash)
	if err != nil {
		log.Error("Failed to insert into auth_references", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO users_and_providers (user_id, provider_id, user_provider_id) VALUES ($1, $2, $3)", u.ID, u.ProviderID, u.UserProviderID)
	if err != nil {
		log.Error("Failed to insert into users_and_providers", zap.Error(err))
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

	log.Debug("Manual user is added")
	return err
}

func (pa *PostgresAuth) GetUser(ctx context.Context, log *zap.Logger, providerID int, userProviderID string) (u UserInfo, err error) {
	log.Debug("PostgresAuth.GetUser", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", pa.Pool == nil), zap.Int("providerID", providerID), zap.String("userProviderID", userProviderID))
	if pa.Pool == nil || ctx == nil || providerID == 0 || userProviderID == "" {
		log.Error("Invalid input")
		return UserInfo{}, errors.New("invalid input")
	}

	tx, err := pa.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return u, err
	}

	err = tx.QueryRow(ctx, "SELECT user_id FROM users_and_providers WHERE provider_id = $1 AND user_provider_id = $2", providerID, userProviderID).Scan(&u.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, nil
	}
	if err != nil {
		log.Error("Failed to select from users_and_providers", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return u, err
	}

	err = tx.QueryRow(ctx, "SELECT email, name, picture_url FROM users WHERE id = $1", u.ID).Scan(&u.Email, &u.Name, &u.Picture)
	if err != nil {
		log.Error("Failed to select from users", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return u, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return u, err
	}

	log.Debug("User info retrieved", zap.Any("u", u))
	return u, err
}

func (pa *PostgresAuth) AddNewProviderUser(ctx context.Context, log *zap.Logger, u UserInfo) (err error) {
	log.Debug("PostgresAuth.AddNewProviderUser", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", pa.Pool == nil), zap.Any("u", u))
	if pa.Pool == nil || ctx == nil {
		log.Error("Invalid input")
		return errors.New("invalid input")
	}

	tx, err := pa.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO users (id, email, name, picture_url) VALUES ($1, $2, $3, $4)", u.ID, u.Email, u.Name, u.Picture)
	if err != nil {
		log.Error("Failed to insert into users", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO users_and_providers (user_id, provider_id, user_provider_id) VALUES ($1, $2, $3)", u.ID, u.ProviderID, u.UserProviderID)
	if err != nil {
		log.Error("Failed to insert into users_and_providers", zap.Error(err))
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

	log.Debug("Added new user", zap.String("user_id", u.ID))
	return err
}

func (pa *PostgresAuth) GetPasswordHash(ctx context.Context, log *zap.Logger, login string) (hash []byte, err error) {
	log.Debug("PostgresAuth.GetPasswordHash", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", pa.Pool == nil), zap.String("login", login))
	if pa.Pool == nil || ctx == nil || login == "" {
		log.Error("Invalid input")
		return nil, errors.New("invalid input")
	}

	err = pa.QueryRow(ctx, "SELECT password_hash FROM auth_references WHERE user_id = $1", login).Scan(&hash)
	if err != nil {
		log.Error("Failed to query password hash", zap.Error(err))
		return hash, err
	}

	log.Debug("Password hash retrieved")
	return hash, err
}

func (pa *PostgresAuth) SaveRefreshToken(ctx context.Context, log *zap.Logger, login, refreshHash string) (err error) {
	log.Debug("PostgresAuth.SaveRefreshToken", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", pa.Pool == nil), zap.String("login", login), zap.Bool("refreshHash_is_empty", refreshHash == ""))
	if pa.Pool == nil || ctx == nil || login == "" || refreshHash == "" {
		log.Error("Invalid input")
		return errors.New("invalid input")
	}

	tx, err := pa.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}

	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", login).Scan(&count)
	if err != nil {
		log.Error("Failed to count users", zap.Error(err))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	if count == 0 {
		log.Error("User not found")
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return errors.New("user not found")
	}

	_, err = tx.Exec(ctx, "INSERT INTO auth (user_id, refresh_token) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET refresh_token = $2", login, refreshHash)
	if err != nil {
		log.Error("Failed to update auth", zap.Error(err))
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

	log.Debug("Refresh token saved")
	return err
}

func (pa *PostgresAuth) GetRefreshToken(ctx context.Context, log *zap.Logger, login string) (refreshHash string, err error) {
	log.Debug("PostgresAuth.GetRefreshToken", zap.Bool("ctx_is_nil", ctx == nil), zap.Bool("pool_is_nil", pa.Pool == nil), zap.String("login", login))
	if pa.Pool == nil || ctx == nil || login == "" {
		log.Error("Invalid input")
		return "", errors.New("invalid input")
	}

	err = pa.QueryRow(ctx, "SELECT refresh_token FROM auth WHERE user_id = $1", login).Scan(&refreshHash)
	if err != nil {
		log.Error("Failed to get refresh token", zap.Error(err))
		return refreshHash, err
	}

	log.Debug("Refresh token retrieved", zap.Bool("refreshHash_is_empty", refreshHash == ""))
	return refreshHash, err
}
