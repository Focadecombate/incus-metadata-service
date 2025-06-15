CREATE TABLE IF NOT EXISTS vendor_data (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  data JSONB
);

-- Index for faster lookups by name
CREATE INDEX IF NOT EXISTS idx_vendor_data_name ON vendor_data(name);

-- Index for only one active record per name
CREATE UNIQUE INDEX IF NOT EXISTS idx_vendor_data_active_name ON vendor_data(name, deleted_at)
WHERE
  deleted_at IS NULL;