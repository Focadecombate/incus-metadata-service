package types

// network:
//   version: 2
//   ethernets:
//     # opaque ID for physical interfaces, only referred to by other stanzas
//     id0:
//       match:
//         macaddress: '00:11:22:33:44:55'
//       wakeonlan: true
//       dhcp4: true
//       addresses:
//         - 192.168.14.2/24
//         - 2001:1::1/64
//       gateway4: 192.168.14.1
//       gateway6: 2001:1::2
//       nameservers:
//         search: [foo.local, bar.local]
//         addresses: [8.8.8.8]
//       # static routes
//       routes:
//         - to: 192.0.2.0/24
//           via: 11.0.0.1
//           metric: 3
//     lom:
//       match:
//         driver: ixgbe
//       # you are responsible for setting tight enough match rules
//       # that only match one device if you use set-name
//       set-name: lom1
//       dhcp6: true
//     switchports:
//       # all cards on second PCI bus; unconfigured by themselves, will be added
//       # to br0 below
//       match:
//         name: enp2*
//       mtu: 1280
//   bonds:
//     bond0:
//       interfaces: [id0, lom]
//   bridges:
//     # the key name is the name for virtual (created) interfaces; no match: and
//     # set-name: allowed
//     br0:
//       # IDs of the components; switchports expands into multiple interfaces
//       interfaces: [wlp1s0, switchports]
//       dhcp4: true
//   vlans:
//     en-intra:
//       id: 1
//       link: id0
//       dhcp4: yes

type Match struct {
	MacAddress string `json:"macaddress,omitempty" yaml:"macaddress,omitempty"`
	Driver     string `json:"driver,omitempty" yaml:"driver,omitempty"`
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
}
type Nameservers struct {
	Search    []string `json:"search,omitempty" yaml:"search,omitempty"`
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
}
type Route struct {
	To     string `json:"to,omitempty" yaml:"to,omitempty"`
	Via    string `json:"via,omitempty" yaml:"via,omitempty"`
	Metric int    `json:"metric,omitempty" yaml:"metric,omitempty"`
}

type Ethernet struct {
	Match       *Match       `json:"match,omitempty" yaml:"match,omitempty"`
	WakeOnLan   bool         `json:"wakeonlan,omitempty" yaml:"wakeonlan,omitempty"`
	DHCP4       bool         `json:"dhcp4,omitempty" yaml:"dhcp4,omitempty"`
	Addresses   []string     `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Gateway4    string       `json:"gateway4,omitempty" yaml:"gateway4,omitempty"`
	Gateway6    string       `json:"gateway6,omitempty" yaml:"gateway6,omitempty"`
	Nameservers *Nameservers `json:"nameservers,omitempty" yaml:"nameservers,omitempty"`
	Routes      []Route      `json:"routes,omitempty" yaml:"routes,omitempty"`
}

// NetworkConfig is netplan / cloud-init Networking-Config v2. Version is the
// document-root key and is always emitted as an int (2).
type NetworkConfig struct {
	Version   int                 `json:"version" yaml:"version"`
	Ethernets map[string]Ethernet `json:"ethernets" yaml:"ethernets"`
}
