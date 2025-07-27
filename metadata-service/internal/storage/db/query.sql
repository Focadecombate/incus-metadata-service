-- name: CreateVendorData :one
INSERT INTO
  vendor_data (name, description, data)
VALUES
  (?, ?, ?) RETURNING *;

-- name: GetVendorData :one
SELECT
  id,
  name,
  description,
  created_at,
  updated_at,
  data
FROM
  vendor_data
WHERE
  name = ?
  and deleted_at IS NULL;

-- name: UpdateVendorData :one
UPDATE
  vendor_data
SET
  description = ?,
  data = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = ? RETURNING *;

-- name: DeleteVendorData :exec
UPDATE
  vendor_data
SET
  deleted_at = CURRENT_TIMESTAMP
WHERE
  id = ?;

-- ===== INSTANCES QUERIES =====
-- name: CreateInstance :one
INSERT INTO
  instances (name, project, ip_address)
VALUES
  (?, ?, ?) RETURNING *;

-- name: GetInstance :one
SELECT
  *
FROM
  instances
WHERE
  name = ?
  AND project = ?
  AND deleted_at IS NULL;

-- name: GetInstanceByID :one
SELECT
  *
FROM
  instances
WHERE
  id = ?
  AND deleted_at IS NULL;

-- name: GetInstanceByIP :one
SELECT
  *
FROM
  instances
WHERE
  ip_address = ?
  AND deleted_at IS NULL;

-- name: ListInstances :many
SELECT
  *
FROM
  instances
WHERE
  deleted_at IS NULL
ORDER BY
  created_at DESC;

-- name: ListInstancesByProject :many
SELECT
  *
FROM
  instances
WHERE
  project = ?
  AND deleted_at IS NULL
ORDER BY
  created_at DESC;

-- name: UpdateInstance :one
UPDATE
  instances
SET
  ip_address = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = ? RETURNING *;

-- name: UpdateInstanceIP :exec
UPDATE
  instances
SET
  ip_address = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = ?;

-- name: DeleteInstance :exec
UPDATE
  instances
SET
  deleted_at = CURRENT_TIMESTAMP
WHERE
  id = ?;

-- name: HardDeleteInstance :exec
DELETE FROM
  instances
WHERE
  id = ?;

-- ===== INSTANCE STATE QUERIES =====
-- name: CreateOrUpdateInstanceState :one
INSERT INTO
  instance_state (instance_id, status, status_code, updated_at)
VALUES
  (?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(instance_id) DO
UPDATE
SET
  status = excluded.status,
  status_code = excluded.status_code,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: GetInstanceState :one
SELECT
  *
FROM
  instance_state
WHERE
  instance_id = ?;

-- name: DeleteInstanceState :exec
DELETE FROM
  instance_state
WHERE
  instance_id = ?;

-- ===== INSTANCE LOGS QUERIES =====
-- name: CreateInstanceLog :one
INSERT INTO
  instance_logs (instance_id, log_type, level, message)
VALUES
  (?, ?, ?, ?) RETURNING *;

-- name: GetInstanceLogs :many
SELECT
  *
FROM
  instance_logs
WHERE
  instance_id = ?
ORDER BY
  created_at DESC
LIMIT
  ? OFFSET ?;

-- name: GetInstanceLogsByType :many
SELECT
  *
FROM
  instance_logs
WHERE
  instance_id = ?
  AND log_type = ?
ORDER BY
  created_at DESC
LIMIT
  ? OFFSET ?;

-- name: GetInstanceLogsByLevel :many
SELECT
  *
FROM
  instance_logs
WHERE
  instance_id = ?
  AND level = ?
ORDER BY
  created_at DESC
LIMIT
  ? OFFSET ?;

-- name: DeleteInstanceLogs :exec
DELETE FROM
  instance_logs
WHERE
  instance_id = ?;

-- name: DeleteOldInstanceLogs :exec
DELETE FROM
  instance_logs
WHERE
  created_at < ?
  AND instance_id = ?;

-- ===== PROFILES QUERIES =====
-- name: CreateProfile :one
INSERT INTO
  profiles (name, project)
VALUES
  (?, ?) RETURNING *;

-- name: GetProfile :one
SELECT
  *
FROM
  profiles
WHERE
  name = ?
  AND project = ?
  AND deleted_at IS NULL;

-- name: ListProfiles :many
SELECT
  *
FROM
  profiles
WHERE
  deleted_at IS NULL
ORDER BY
  created_at DESC;

-- name: ListProfilesByProject :many
SELECT
  *
FROM
  profiles
WHERE
  project = ?
  AND deleted_at IS NULL
ORDER BY
  created_at DESC;

-- name: UpdateProfile :one
UPDATE
  profiles
SET
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = ? RETURNING *;

-- name: DeleteProfile :exec
UPDATE
  profiles
SET
  deleted_at = CURRENT_TIMESTAMP
WHERE
  id = ?;