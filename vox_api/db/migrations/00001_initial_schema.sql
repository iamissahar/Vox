-- +goose Up
CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    email       TEXT NOT NULL,
    name        TEXT NOT NULL,
    picture_url TEXT NOT NULL
);

CREATE TABLE auth_references (
    user_id       TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash BYTEA NOT NULL
);

CREATE TABLE auth (
    user_id       TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT
);

CREATE TABLE providers (
    id   INT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE users_and_providers (
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id      INT  NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    user_provider_id TEXT NOT NULL,
    PRIMARY KEY (provider_id, user_provider_id)
);

CREATE TABLE files (
    id          TEXT PRIMARY KEY,
    full_path    TEXT NOT NULL,
    type        TEXT NOT NULL,
    text        TEXT NOT NULL
);

CREATE TABLE files_and_users (
    file_id TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (file_id, user_id)
);

CREATE TABLE user_voice (
    user_id  TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    text     TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS user_voice;
DROP TABLE IF EXISTS files_and_users;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS users_and_providers;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS auth;
DROP TABLE IF EXISTS auth_references;
DROP TABLE IF EXISTS users;
