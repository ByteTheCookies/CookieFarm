-- name: GetFlagByCode :one
SELECT *
FROM flags
WHERE flag_code = ?
LIMIT 1;

-- name: GetFlagsByTeam :many
SELECT *
FROM flags
WHERE team_id = ?
ORDER BY submit_time DESC
LIMIT ? OFFSET ?;

-- name: GetAllFlags :many
SELECT *
FROM flags
ORDER BY submit_time DESC;

-- name: CountFlags :one
SELECT COUNT(*)
FROM flags
LIMIT 1;

-- name: GetFirstNFlags :many
SELECT *
FROM flags
ORDER BY submit_time DESC
LIMIT ?;

-- name: GetUnsubmittedFlags :many
SELECT *
FROM flags
WHERE status = 'UNSUBMITTED'
ORDER BY submit_time ASC
LIMIT ?;

-- name: GetPagedFlags :many
SELECT *
FROM flags
ORDER BY submit_time DESC
LIMIT ? OFFSET ?;

-- name: GetAllFlagCodes :many
SELECT flag_code FROM flags;

-- name: GetFirstNFlagCodes :many
SELECT flag_code FROM flags
LIMIT ?;

-- name: GetUnsubmittedFlagCodes :many
SELECT flag_code FROM flags
WHERE status = 'UNSUBMITTED'
LIMIT ?;

-- name: GetFilteredFlags :many
SELECT * FROM flags
WHERE
    (team_id = sqlc.narg('team_id') OR sqlc.narg('team_id') IS NULL)
    AND (status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
    AND (
        sqlc.narg('search') IS NULL
        OR (
            (sqlc.narg('search_field') = 'flag_code'    AND flag_code    LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'service_name' AND service_name LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'exploit_name' AND exploit_name LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'msg'          AND msg          LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'all' AND (
                flag_code    LIKE sqlc.narg('search_like')
                OR service_name  LIKE sqlc.narg('search_like')
                OR exploit_name  LIKE sqlc.narg('search_like')
                OR msg           LIKE sqlc.narg('search_like')
                OR CAST(team_id AS TEXT) LIKE sqlc.narg('search_like')
            ))
            OR (sqlc.narg('search_field') IS NULL AND flag_code LIKE sqlc.narg('search_like'))
        )
)
ORDER BY submit_time DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');


-- name: CountFilteredFlags :one
SELECT COUNT(*) FROM flags
WHERE
    (team_id = sqlc.narg('team_id') OR sqlc.narg('team_id') IS NULL)
    AND (status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
    AND (service_name = sqlc.narg('service_name') OR sqlc.narg('service_name') IS NULL)
    AND (
        sqlc.narg('search') IS NULL
        OR (
            (sqlc.narg('search_field') = 'flag_code'    AND flag_code    LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'service_name' AND service_name LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'exploit_name' AND exploit_name LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'msg'          AND msg          LIKE sqlc.narg('search_like'))
            OR (sqlc.narg('search_field') = 'all' AND (
                flag_code    LIKE sqlc.narg('search_like')
                OR service_name  LIKE sqlc.narg('search_like')
                OR exploit_name  LIKE sqlc.narg('search_like')
                OR msg           LIKE sqlc.narg('search_like')
                OR CAST(team_id AS TEXT) LIKE sqlc.narg('search_like')
            ))
            OR (sqlc.narg('search_field') IS NULL AND flag_code LIKE sqlc.narg('search_like'))
    )
);

-- name: GetPagedFlagCodes :many
SELECT flag_code FROM flags
LIMIT ? OFFSET ?;

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
WHERE flag_code = ?;

-- name: DeleteFlagByCode :exec
DELETE FROM flags
WHERE flag_code = ?;

-- name: DeleteFlagByTTL :execrows
DELETE FROM flags
WHERE response_time < (CAST(strftime('%s', 'now') AS INTEGER) - ?);
