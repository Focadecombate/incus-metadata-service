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
  instances (name, project, source_node, ip_address)
VALUES
  (?, ?, ?, ?) ON CONFLICT(name, project) DO
UPDATE
SET
  ip_address = excluded.ip_address,
  source_node = excluded.source_node,
  deleted_at = NULL,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

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

-- name: ListActiveInstancesBySourceNode :many
SELECT
  *
FROM
  instances
WHERE
  source_node = ?
  AND deleted_at IS NULL;

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

-- ===== INSTANCE METADATA QUERIES =====
-- name: CreateOrUpdateInstanceMetadata :one
INSERT INTO
  instance_metadata (instance_id, metadata)
VALUES
  (?, ?) ON CONFLICT(instance_id) DO
UPDATE
SET
  metadata = excluded.metadata,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: GetInstanceMetadata :one
SELECT
  *
FROM
  instance_metadata
WHERE
  instance_id = ?;

-- name: GetInstanceMetadataByIP :one
SELECT
  im.*
FROM
  instance_metadata im
  JOIN instances i ON im.instance_id = i.id
WHERE
  i.ip_address = ?
  AND i.deleted_at IS NULL;

-- name: DeleteInstanceMetadata :exec
DELETE FROM
  instance_metadata
WHERE
  instance_id = ?;

-- ===== INSTANCE USER DATA QUERIES =====
-- name: CreateOrUpdateInstanceUserData :one
INSERT INTO
  instance_user_data (instance_id, user_data)
VALUES
  (?, ?) ON CONFLICT(instance_id) DO
UPDATE
SET
  user_data = excluded.user_data,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: GetInstanceUserData :one
SELECT
  *
FROM
  instance_user_data
WHERE
  instance_id = ?;

-- name: GetInstanceUserDataByIP :one
SELECT
  iud.*
FROM
  instance_user_data iud
  JOIN instances i ON iud.instance_id = i.id
WHERE
  i.ip_address = ?
  AND i.deleted_at IS NULL;

-- name: DeleteInstanceUserData :exec
DELETE FROM
  instance_user_data
WHERE
  instance_id = ?;

-- ===== INSTANCE VENDOR DATA QUERIES =====
-- name: CreateOrUpdateInstanceVendorData :one
INSERT INTO
  instance_vendor_data (instance_id, vendor_data)
VALUES
  (?, ?) ON CONFLICT(instance_id) DO
UPDATE
SET
  vendor_data = excluded.vendor_data,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: GetInstanceVendorData :one
SELECT
  *
FROM
  instance_vendor_data
WHERE
  instance_id = ?;

-- name: GetInstanceVendorDataByIP :one
SELECT
  ivd.*
FROM
  instance_vendor_data ivd
  JOIN instances i ON ivd.instance_id = i.id
WHERE
  i.ip_address = ?
  AND i.deleted_at IS NULL;

-- name: DeleteInstanceVendorData :exec
DELETE FROM
  instance_vendor_data
WHERE
  instance_id = ?;

-- ===== INSTANCE NETWORK CONFIG QUERIES =====
-- name: CreateOrUpdateInstanceNetworkConfig :one
INSERT INTO
  instance_network_config (instance_id, network_config)
VALUES
  (?, ?) ON CONFLICT(instance_id) DO
UPDATE
SET
  network_config = excluded.network_config,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: GetInstanceNetworkConfig :one
SELECT
  *
FROM
  instance_network_config
WHERE
  instance_id = ?;

-- name: GetInstanceNetworkConfigByIP :one
SELECT
  inc.*
FROM
  instance_network_config inc
  JOIN instances i ON inc.instance_id = i.id
WHERE
  i.ip_address = ?
  AND i.deleted_at IS NULL;

-- name: DeleteInstanceNetworkConfig :exec
DELETE FROM
  instance_network_config
WHERE
  instance_id = ?;