package model

type Rack struct {
	ID           string `json:"id"`
	RackCidr     string `json:"rack_cidr"`
	RackCidrGw   string `json:"rack_cidr_gw"`
	ProviderCidr string `json:"provider_cidr"`
	RackTag      string `json:"rack_tag"`
}
