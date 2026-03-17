-- +goose Up
INSERT INTO providers (id, name) VALUES
    (-1, 'google'),
    (-2, 'github'),
    (-3, 'vox')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM providers WHERE id IN (-1, -2, -3);
