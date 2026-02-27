-- name: GetFlagByCode :one
SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name
FROM flags
WHERE flag_code = ?
LIMIT 1;

-- name: ListFlagsByTeam :many
SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name
FROM flags
WHERE team_id = ?
ORDER BY submit_time DESC
LIMIT ? OFFSET ?;

-- name: ListAllFlags :many
SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name
FROM flags
ORDER BY submit_time DESC;

-- name: ListFirstNFlags :many
SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name
FROM flags
ORDER BY submit_time DESC
LIMIT ?;

-- name: ListUnsubmittedFlags :many
SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name
FROM flags
WHERE status = 'UNSUBMITTED'
ORDER BY submit_time ASC
LIMIT ?;

-- name: ListPagedFlags :many
SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name
FROM flags
ORDER BY submit_time DESC
LIMIT ? OFFSET ?;

-- name: ListAllFlagCodes :many
SELECT flag_code FROM flags;

-- name: ListFirstNFlagCodes :many
SELECT flag_code FROM flags
LIMIT ?;

-- name: ListUnsubmittedFlagCodes :many
SELECT flag_code FROM flags
WHERE status = 'UNSUBMITTED'
LIMIT ?;

-- name: ListPagedFlagCodes :many
SELECT flag_code FROM flags
LIMIT ? OFFSET ?;

-- name: AddFlags :exec
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

-- name: DeleteFlagByTTL :exec
DELETE FROM flags
WHERE response_time < strftime('%s', 'now', ?);
