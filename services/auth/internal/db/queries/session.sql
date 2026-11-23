-- name: CreateSession :one
INSERT INTO "session" (id, token, user_id, expires_at, ip_address, user_agent, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
RETURNING *;

-- name: GetUserBySessionToken :one
SELECT sqlc.embed(u)
FROM "session" s
JOIN "user" u ON u.id = s.user_id
WHERE s.token = $1 AND s.expires_at > now();

-- name: DeleteSessionByToken :exec
DELETE FROM "session" WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM "session" WHERE expires_at <= now();
