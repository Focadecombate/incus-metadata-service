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