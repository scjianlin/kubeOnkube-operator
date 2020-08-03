package model

// 机柜CIDR
type Rack struct {
	ID           string      `json:"id"`
	RackCidr     string      `json:"rackCidr"`
	RackCidrGw   string      `json:"rackCidrGw"`
	ProviderCidr string      `json:"providerCidr"`
	RackTag      string      `json:"rackTag"`
	IsMaster     int         `json:"isMaster"` //值为0表示False,值为1表示True
	HostAddr     []*HostAddr `json:"hostAddr"`
	PodCidr      []*PodAddr  `json:"podCidr"`
	PodNum       int         `json:"podNum"`
}

// 主机地址
type HostAddr struct {
	ID        string   `json:"id"`
	IPADDR    string   `json:"ipAddr"`
	NetMask   string   `json:"netMask"`
	GateWay   string   `json:"gateWay"`
	DnsServer []string `json:"dnsServer"`
}

// POD 地址段
type PodAddr struct {
	ID           string `json:"id"`
	RangeStart   string `json:"rangeStart"`
	RangeEnd     string `json:"rangeEnd"`
	DefaultRoute string `json:"defaultRoute"`
}

//
type PodAddrList struct {
	ID           string `json:"id"`
	RangeStart   string `json:"rangeStart"`
	RangeEnd     string `json:"rangeEnd"`
	DefaultRoute string `json:"defaultRoute"`
	RackCidr     string `json:"rackCidr"`
	RackTag      string `json:"rackTag"`
}
