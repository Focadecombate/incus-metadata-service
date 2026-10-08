package events

import (
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
	incus "github.com/lxc/incus/shared/api"
)

func managedBridge(network string) (netGateway, bool) {
	if network == "incusbr0" {
		return netGateway{IPv4: "10.10.10.1", IPv6: "fd42::1"}, true
	}
	return netGateway{}, false
}

func TestBuildNetworkConfigStaticOnManagedNetwork(t *testing.T) {
	ifaces := []ifaceInfo{{
		Name: "eth0", Hwaddr: "00:16:3e:00:00:01",
		IPv4: "10.10.10.39", Netmask: "24",
		IPv6: "fd42::216:3eff:fe00:1", Netmask6: "64",
	}}
	devices := map[string]map[string]string{"eth0": {"type": "nic", "network": "incusbr0"}}

	cfg := buildNetworkConfig(ifaces, devices, managedBridge)

	if cfg.Version != 2 {
		t.Fatalf("version = %d, want 2", cfg.Version)
	}
	eth, ok := cfg.Ethernets["eth0"]
	if !ok {
		t.Fatalf("ethernets keyed by interface name, got %v", cfg.Ethernets)
	}
	if eth.Match != nil {
		t.Errorf("match must be absent (PermanentMACAddress never matches a veth), got %+v", eth.Match)
	}
	if eth.DHCP4 {
		t.Errorf("static config must not enable dhcp4")
	}
	wantAddrs := []string{"10.10.10.39/24", "fd42::216:3eff:fe00:1/64"}
	if len(eth.Addresses) != 2 || eth.Addresses[0] != wantAddrs[0] || eth.Addresses[1] != wantAddrs[1] {
		t.Errorf("addresses = %v, want %v", eth.Addresses, wantAddrs)
	}
	wantRoutes := []types.Route{{To: "0.0.0.0/0", Via: "10.10.10.1"}, {To: "::/0", Via: "fd42::1"}}
	if len(eth.Routes) != 2 || eth.Routes[0] != wantRoutes[0] || eth.Routes[1] != wantRoutes[1] {
		t.Errorf("routes = %v, want %v", eth.Routes, wantRoutes)
	}
	if eth.Nameservers == nil || len(eth.Nameservers.Addresses) != 1 || eth.Nameservers.Addresses[0] != "10.10.10.1" {
		t.Errorf("nameservers = %+v, want gateway 10.10.10.1", eth.Nameservers)
	}
}

func TestBuildNetworkConfigOmitsIPv6WithoutPrefix(t *testing.T) {
	ifaces := []ifaceInfo{{Name: "eth0", IPv4: "10.10.10.39", Netmask: "24", IPv6: "fd42::1"}}
	devices := map[string]map[string]string{"eth0": {"network": "incusbr0"}}

	eth := buildNetworkConfig(ifaces, devices, managedBridge).Ethernets["eth0"]

	if len(eth.Addresses) != 1 || eth.Addresses[0] != "10.10.10.39/24" {
		t.Errorf("addresses = %v, want only the IPv4 CIDR", eth.Addresses)
	}
	if len(eth.Routes) != 1 {
		t.Errorf("routes = %v, want only the IPv4 default route", eth.Routes)
	}
}

func TestBuildNetworkConfigFallsBackToDHCP(t *testing.T) {
	cases := map[string]struct {
		iface   ifaceInfo
		devices map[string]map[string]string
	}{
		"unmanaged or unknown network": {
			iface:   ifaceInfo{Name: "eth0", IPv4: "192.168.1.5", Netmask: "24"},
			devices: map[string]map[string]string{"eth0": {"nictype": "bridged", "parent": "br0"}},
		},
		"no device entry": {
			iface:   ifaceInfo{Name: "eth0", IPv4: "10.10.10.5", Netmask: "24"},
			devices: nil,
		},
		"no IPv4 yet": {
			iface:   ifaceInfo{Name: "eth0"},
			devices: map[string]map[string]string{"eth0": {"network": "incusbr0"}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eth := buildNetworkConfig([]ifaceInfo{tc.iface}, tc.devices, managedBridge).Ethernets["eth0"]
			if !eth.DHCP4 || len(eth.Addresses) != 0 || len(eth.Routes) != 0 || eth.Match != nil {
				t.Errorf("want plain dhcp4 fallback, got %+v", eth)
			}
		})
	}
}

func TestExtractInterfaceInfoCapturesIPv6Prefix(t *testing.T) {
	state := &incus.InstanceState{Network: map[string]incus.InstanceStateNetwork{
		"lo": {Addresses: []incus.InstanceStateNetworkAddress{{Family: "inet", Address: "127.0.0.1", Netmask: "8", Scope: "local"}}},
		"eth0": {Hwaddr: "00:16:3e:00:00:01", Addresses: []incus.InstanceStateNetworkAddress{
			{Family: "inet6", Address: "fe80::1", Netmask: "64", Scope: "link"},
			{Family: "inet", Address: "10.10.10.39", Netmask: "24", Scope: "global"},
			{Family: "inet6", Address: "fd42::39", Netmask: "64", Scope: "global"},
		}},
	}}

	ifaces := extractInterfaceInfo(state)

	if len(ifaces) != 1 {
		t.Fatalf("want 1 non-loopback interface, got %d", len(ifaces))
	}
	got := ifaces[0]
	want := ifaceInfo{Name: "eth0", Hwaddr: "00:16:3e:00:00:01", IPv4: "10.10.10.39", Netmask: "24", IPv6: "fd42::39", Netmask6: "64"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestIPFromCIDR(t *testing.T) {
	cases := map[string]string{"10.10.10.1/24": "10.10.10.1", "fd42::1/64": "fd42::1", "auto": "", "none": "", "": ""}
	for in, want := range cases {
		if got := ipFromCIDR(in); got != want {
			t.Errorf("ipFromCIDR(%q) = %q, want %q", in, got, want)
		}
	}
}
