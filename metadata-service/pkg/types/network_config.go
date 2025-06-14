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
	MacAddress string `json:"macaddress" yaml:"macaddress"`
	Driver     string `json:"driver" yaml:"driver"`
	Name       string `json:"name" yaml:"name"`
}
type Nameservers struct {
	Search    []string `json:"search" yaml:"search"`
	Addresses []string `json:"addresses" yaml:"addresses"`
}
type Route struct {
	To     string `json:"to" yaml:"to"`
	Via    string `json:"via" yaml:"via"`
	Metric int    `json:"metric" yaml:"metric"`
}

type Ethernet struct {
	Version     int         `json:"version" yaml:"version"`
	Match       Match       `json:"match" yaml:"match"`
	WakeOnLan   bool        `json:"wakeonlan" yaml:"wakeonlan"`
	DHCP4       bool        `json:"dhcp4" yaml:"dhcp4"`
	Addresses   []string    `json:"addresses" yaml:"addresses"`
	Gateway4    string      `json:"gateway4" yaml:"gateway4"`
	Gateway6    string      `json:"gateway6" yaml:"gateway6"`
	Nameservers Nameservers `json:"nameservers" yaml:"nameservers"`
	Routes      []Route     `json:"routes" yaml:"routes"`
}

type NetworkConfig struct {
	Version   int                 `json:"version" yaml:"version"`
	Ethernets map[string]Ethernet `json:"ethernets" yaml:"ethernets"`
}
