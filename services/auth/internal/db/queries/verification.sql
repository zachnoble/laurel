-- name: CreateVerification :one
INSERT INTO "verification" (id, identifier, value, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestVerificationByIdentifier :one
SELECT * FROM "verification"
WHERE identifier = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteVerificationsByIdentifier :exec
DELETE FROM "verification" WHERE identifier = $1;

-- name: IncrementVerificationAttempts :exec
UPDATE "verification" SET attempts = attempts + 1 WHERE id = $1;

-- name: DeleteExpiredVerifications :exec
DELETE FROM "verification" WHERE expires_at <= now();
