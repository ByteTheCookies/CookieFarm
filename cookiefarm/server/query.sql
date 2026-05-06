-- name: GetFlagByCode :one
SELECT * FROM flags
WHERE flag_code = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetAllFlags :many
SELECT *
FROM flags
WHERE deleted_at IS NULL
ORDER BY submit_time DESC;

-- name: CountFlags :one
SELECT COUNT(*)
FROM flags
WHERE deleted_at IS NULL
LIMIT 1;

-- name: GetFirstNFlags :many
SELECT *
FROM flags
WHERE deleted_at IS NULL
ORDER BY submit_time DESC
LIMIT ?;

-- name: GetUnsubmittedFlags :many
SELECT *
FROM flags
WHERE status = 0
AND deleted_at IS NULL
ORDER BY submit_time ASC
LIMIT ?;

-- name: GetAllFlagCodes :many
SELECT flag_code FROM flags WHERE deleted_at IS NULL;

-- name: GetFirstNFlagCodes :many
SELECT flag_code FROM flags
WHERE deleted_at IS NULL
ORDER BY submit_time DESC
LIMIT ?;

-- name: GetUnsubmittedFlagCodes :many
SELECT flag_code FROM flags
WHERE status = 0
AND deleted_at IS NULL
LIMIT ?;

-- name: AddFlag :exec
INSERT OR IGNORE INTO flags(
	flag_code, service_name, port_service,
	submit_time, response_time, status,
	team_id, msg, username, exploit_name
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateFlagStatusByCode :exec
UPDATE flags
SET
	status = ?,
	msg = ?,
	response_time = ?
WHERE flag_code = ?
AND deleted_at IS NULL;

-- name: DeleteFlagByCode :exec
UPDATE flags
SET deleted_at = CAST(strftime('%s', 'now') AS INTEGER)
WHERE flag_code = ?
AND deleted_at IS NULL;

-- name: DeleteFlagByTTL :execrows
UPDATE flags
SET deleted_at = CAST(strftime('%s', 'now') AS INTEGER)
WHERE response_time < (CAST(strftime('%s', 'now') AS INTEGER) - ?)
AND deleted_at IS NULL;

-- name: FlagsStats :many
SELECT
    team_id,
    COUNT(*) AS total_flags,
    SUM(status = 1) AS accepted_flags,
    SUM(status = 2) AS denied_flags,
    SUM(status = 0) AS unsubmitted_flags,
    SUM(status = 3) AS error_flags,
    SUM(status = 4) AS not_valid_flags
FROM flags
WHERE deleted_at IS NULL
GROUP BY team_id
ORDER BY team_id ASC;

-- name: FlagsTickStats :many
SELECT
    submit_time / ? AS bucket,
    COUNT(*) AS total,
    SUM(status = 0) AS queued,
    SUM(status = 1) AS accepted,
    SUM(status = 2) AS denied,
    SUM(status = 3) AS error,
    SUM(status = 4) AS invalid
FROM flags
WHERE submit_time > 0
AND deleted_at IS NULL
GROUP BY bucket
ORDER BY bucket ASC;

-- name: FlagsExploitShare :many
SELECT
	exploit_name,
	COUNT(*) AS value
FROM flags
WHERE exploit_name IS NOT NULL
GROUP BY exploit_name
ORDER BY value DESC, exploit_name ASC;

-- name: FlagsExploitTickStats :many
SELECT
	exploit_name,
	submit_time / ? AS bucket,
	COUNT(*) AS value
FROM flags
WHERE submit_time > 0
AND deleted_at IS NULL
AND exploit_name IS NOT NULL
GROUP BY exploit_name, bucket
ORDER BY exploit_name ASC, bucket ASC;

-- name: FlagsExploitStatusStats :many
SELECT
	exploit_name,
	COUNT(*) AS total,
	SUM(status = 0) AS queued,
	SUM(status = 1) AS accepted,
	SUM(status = 2) AS denied,
	SUM(status = 3) AS error,
	SUM(status = 4) AS invalid
FROM flags
WHERE deleted_at IS NULL
AND exploit_name IS NOT NULL
GROUP BY exploit_name
ORDER BY total DESC, exploit_name ASC;

-- EXPLOITS QUERIES

-- name: GetExploitByHash :one
SELECT *
FROM exploits
WHERE hash = ?
LIMIT 1;

-- name: GetExploitsByUsername :many
SELECT *
FROM exploits
WHERE username = ? AND id > ?
ORDER BY submit_time DESC
LIMIT ?;

-- name: GetAllExploits :many
SELECT *
FROM exploits
ORDER BY submit_time DESC;

-- name: CountExploits :one
SELECT COUNT(*)
FROM exploits
LIMIT 1;

-- name: GetExploitsByName :many
SELECT *
FROM exploits
WHERE name = ?
ORDER BY submit_time DESC;

-- name: CreateExploit :exec

INSERT INTO exploits(name, hash, submit_time, username, version)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteExploitByID :exec
DELETE FROM exploits
WHERE id = ?;
