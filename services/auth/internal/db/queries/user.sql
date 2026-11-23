-- name: GetUserByID :one
SELECT * FROM "user" WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: CreateUser :one
INSERT INTO "user" (id, name, email, email_verified)
VALUES ($1, $2, $3, $4)
RETURNING *;
