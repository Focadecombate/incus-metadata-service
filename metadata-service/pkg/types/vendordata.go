package types

// VendorData represents the vendor data configuration for cloud-init.
type VendorData struct {
	// VendorName is the name of the vendor providing the data.
	VendorName string `json:"vendor_name" yaml:"vendor_name"`
	// VendorVersion is the version of the vendor data.
	VendorVersion string `json:"vendor_version" yaml:"vendor_version"`
	// VendorData is a map containing vendor-specific data.
	VendorData map[string]any `json:"vendor_data" yaml:"vendor_data"`
}
