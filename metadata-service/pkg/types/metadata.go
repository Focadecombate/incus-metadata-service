package types

// Package types defines the data structures used in the metadata service.
// These structures represent the metadata, network configuration, and other related information
// that the service provides to clients.
// The structures are designed to be serialized to JSON format for easy consumption by clients.
// Yaml tags are used for compatibility with cloud-init and other tools that expect metadata in a specific format.

type Placement struct {
	HostID string `json:"host-id" yaml:"host-id"`
	AvailabilityZone string `json:"availability-zone" yaml:"availability-zone"`
	Region string `json:"region" yaml:"region"`
	Project string `json:"project" yaml:"project"`
}

type Mac struct {
	DeviceNumber string `json:"device-number" yaml:"device-number"`
	LocalHostname string `json:"local-hostname" yaml:"local-hostname"`
	LocalIPv4 string `json:"local-ipv4" yaml:"local-ipv4"`
	LocalIPv6 string `json:"local-ipv6" yaml:"local-ipv6"`
	PublicIPv4 string `json:"public-ipv4" yaml:"public-ipv4"`
	PublicIPv6 string `json:"public-ipv6" yaml:"public-ipv6"`
	Mac string `json:"mac" yaml:"mac"`
}

type Interfaces struct {
	Macs map[string]Mac `json:"macs" yaml:"macs"`
}

type Network struct {
	Interfaces Interfaces `json:"interfaces" yaml:"interfaces"`
}

type Metadata struct {
	InstanceID     string `json:"instance-id" yaml:"instance-id"`
	Hostname       string `json:"hostname" yaml:"hostname"`
	LocalHostname string `json:"local-hostname" yaml:"local-hostname"`
	AvailabilityZone string `json:"availability-zone"`
	Region         string `json:"region" yaml:"region"`
	LocalIPv4      string `json:"local-ipv4" yaml:"local-ipv4"`
	LocalIPv6      string `json:"local-ipv6" yaml:"local-ipv6"`
	PublicIPv4     string `json:"public-ipv4" yaml:"public-ipv4"`
	PublicIPv6     string `json:"public-ipv6" yaml:"public-ipv6"`
	PublicKeys		 []string `json:"public-keys" yaml:"public-keys"`
	SecurityGroups []string `json:"security-groups" yaml:"security-groups"`
	Placement 		Placement `json:"placement" yaml:"placement"`
	Network				Network `json:"network" yaml:"network"`
}