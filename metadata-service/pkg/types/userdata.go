package types

// Package types defines the data structures used in the metadata service.
// These structures represent the user data configuration for cloud-init,
type User struct {
	Name              string   `json:"name" yaml:"name"`
	Sudo              string   `json:"sudo" yaml:"sudo"` 
	Shell             string   `json:"shell" yaml:"shell"`
	SSHAuthorizedKeys []string `json:"ssh_authorized_keys" yaml:"ssh_authorized_keys"`
	Groups            []string `json:"groups" yaml:"groups"`
}

// File represents a file to be written by cloud-init.
// It includes the file path, content, and permissions.
type File struct {
	Path        string `json:"path" yaml:"path"`
	Content     string `json:"content" yaml:"content"`
	Permissions string `json:"permissions" yaml:"permissions"`
}

// UserData represents the user data configuration for cloud-init.
type UserData struct {
	Hostname       string   `json:"hostname" yaml:"hostname"`
	ManageEtcHosts bool     `json:"manage_etc_hosts" yaml:"manage_etc_hosts"`
	Users          []User   `json:"users" yaml:"users"`
	Packages       []string `json:"packages" yaml:"packages"`
	PackageUpdate  bool     `json:"package_update" yaml:"package_update"`
	PackageUpgrade bool     `json:"package_upgrade" yaml:"package_upgrade"`
	WriteFiles     []File   `json:"write_files" yaml:"write_files"`
	RunCommands    []string `json:"runcmd" yaml:"runcmd"`
	FinalMessage   string   `json:"final_message" yaml:"final_message"`
}
