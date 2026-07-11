package types

// Package types defines the data structures used in the metadata service.
// These structures represent the metadata, network configuration, and other related information
// that the service provides to clients.
// The structures are designed to be serialized to JSON format for easy consumption by clients.
// Yaml tags are used for compatibility with cloud-init and other tools that expect metadata in a specific format.

type Placement struct {
	HostID string `json:"host-id" yaml:"host-id,omitempty"`
	AvailabilityZone string `json:"availability-zone" yaml:"availability-zone,omitempty"`
	Region string `json:"region" yaml:"region,omitempty"`
	Project string `json:"project" yaml:"project,omitempty"`
}

type Mac struct {
	DeviceNumber string `json:"device-number" yaml:"device-number,omitempty"`
	LocalHostname string `json:"local-hostname" yaml:"local-hostname,omitempty"`
	LocalIPv4 string `json:"local-ipv4" yaml:"local-ipv4,omitempty"`
	LocalIPv6 string `json:"local-ipv6" yaml:"local-ipv6,omitempty"`
	PublicIPv4 string `json:"public-ipv4" yaml:"public-ipv4,omitempty"`
	PublicIPv6 string `json:"public-ipv6" yaml:"public-ipv6,omitempty"`
	Mac string `json:"mac" yaml:"mac,omitempty"`
}

type Interfaces struct {
	Macs map[string]Mac `json:"macs" yaml:"macs,omitempty"`
}

type Network struct {
	Interfaces Interfaces `json:"interfaces" yaml:"interfaces,omitempty"`
}

type Metadata struct {
	InstanceID     string `json:"instance-id" yaml:"instance-id"`
	Hostname       string `json:"hostname" yaml:"hostname,omitempty"`
	LocalHostname string `json:"local-hostname" yaml:"local-hostname"`
	AvailabilityZone string `json:"availability-zone" yaml:"availability-zone,omitempty"`
	Region         string `json:"region" yaml:"region,omitempty"`
	LocalIPv4      string `json:"local-ipv4" yaml:"local-ipv4,omitempty"`
	LocalIPv6      string `json:"local-ipv6" yaml:"local-ipv6,omitempty"`
	PublicIPv4     string `json:"public-ipv4" yaml:"public-ipv4,omitempty"`
	PublicIPv6     string `json:"public-ipv6" yaml:"public-ipv6,omitempty"`
	PublicKeys		 []string `json:"public-keys" yaml:"public-keys,omitempty"`
	SecurityGroups []string `json:"security-groups" yaml:"security-groups,omitempty"`
	Placement 		Placement `json:"placement" yaml:"placement,omitempty"`
	Network				Network `json:"network" yaml:"network,omitempty"`
}