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

-- Instances table to store VMs/containers created in Incus
CREATE TABLE IF NOT EXISTS instances (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  project TEXT NOT NULL DEFAULT 'default',
  ip_address TEXT, -- IP address for instance identification
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE(name, project)
);

-- Instance state table for current runtime state
CREATE TABLE IF NOT EXISTS instance_state (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  instance_id INTEGER NOT NULL UNIQUE,
  status TEXT NOT NULL,
  status_code INTEGER NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);

-- Instance logs table for operation and event logs
CREATE TABLE IF NOT EXISTS instance_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  instance_id INTEGER NOT NULL,
  log_type TEXT NOT NULL CHECK (log_type IN ('operation', 'event', 'console', 'audit')),
  level TEXT NOT NULL CHECK (level IN ('debug', 'info', 'warn', 'error', 'fatal')),
  message TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);

-- Profiles table to store Incus profiles
CREATE TABLE IF NOT EXISTS profiles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  project TEXT NOT NULL DEFAULT 'default',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE(name, project)
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instances_name ON instances(name);
CREATE INDEX IF NOT EXISTS idx_instances_project ON instances(project);
CREATE INDEX IF NOT EXISTS idx_instances_ip_address ON instances(ip_address);
CREATE INDEX IF NOT EXISTS idx_instances_deleted_at ON instances(deleted_at);
CREATE INDEX IF NOT EXISTS idx_instances_active ON instances(name, project, deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_instance_state_instance_id ON instance_state(instance_id);
CREATE INDEX IF NOT EXISTS idx_instance_state_status ON instance_state(status);
CREATE INDEX IF NOT EXISTS idx_instance_state_updated_at ON instance_state(updated_at);

CREATE INDEX IF NOT EXISTS idx_instance_logs_instance_id ON instance_logs(instance_id);
CREATE INDEX IF NOT EXISTS idx_instance_logs_type ON instance_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_instance_logs_level ON instance_logs(level);
CREATE INDEX IF NOT EXISTS idx_instance_logs_created_at ON instance_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_profiles_name_project ON profiles(name, project);
CREATE INDEX IF NOT EXISTS idx_profiles_deleted_at ON profiles(deleted_at);