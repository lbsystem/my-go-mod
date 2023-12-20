package cloudflare

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	TypeA    = "A"
	TypeAAAA = "AAAA"
	Version  = "v1.11.6"
)

type Zone struct {
	Result []struct {
		Id                  string      `json:"id"`
		Name                string      `json:"name"`
		Status              string      `json:"status"`
		Paused              bool        `json:"paused"`
		Type                string      `json:"type"`
		DevelopmentMode     int         `json:"development_mode"`
		NameServers         []string    `json:"name_servers"`
		OriginalNameServers []string    `json:"original_name_servers"`
		OriginalRegistrar   interface{} `json:"original_registrar"`
		OriginalDnshost     interface{} `json:"original_dnshost"`
		ModifiedOn          string      `json:"modified_on"`
		CreatedOn           time.Time   `json:"created_on"`
		ActivatedOn         time.Time   `json:"activated_on"`
		Meta                struct {
			Step                    int  `json:"step"`
			CustomCertificateQuota  int  `json:"custom_certificate_quota"`
			PageRuleQuota           int  `json:"page_rule_quota"`
			PhishingDetected        bool `json:"phishing_detected"`
			MultipleRailgunsAllowed bool `json:"multiple_railguns_allowed"`
		} `json:"meta"`
		Owner struct {
			Id    string `json:"id"`
			Type  string `json:"type"`
			Email string `json:"email"`
		} `json:"owner"`
		Account struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
		Tenant struct {
			Id   interface{} `json:"id"`
			Name interface{} `json:"name"`
		} `json:"tenant"`
		TenantUnit struct {
			Id interface{} `json:"id"`
		} `json:"tenant_unit"`
		Permissions []string `json:"permissions"`
		Plan        struct {
			Id                string `json:"id"`
			Name              string `json:"name"`
			Price             int    `json:"price"`
			Currency          string `json:"currency"`
			Frequency         string `json:"frequency"`
			IsSubscribed      bool   `json:"is_subscribed"`
			CanSubscribe      bool   `json:"can_subscribe"`
			LegacyId          string `json:"legacy_id"`
			LegacyDiscount    bool   `json:"legacy_discount"`
			ExternallyManaged bool   `json:"externally_managed"`
		} `json:"plan"`
	} `json:"result"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

// T 为cloudflare records 回复的结构体
type T struct {
	Result []struct {
		Id        string `json:"id"`
		ZoneId    string `json:"zone_id"`
		ZoneName  string `json:"zone_name"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Content   string `json:"content"`
		Proxiable bool   `json:"proxiable"`
		Proxied   bool   `json:"proxied"`
		Ttl       int    `json:"ttl"`
		Locked    bool   `json:"locked"`
		Meta      struct {
			AutoAdded           bool   `json:"auto_added"`
			ManagedByApps       bool   `json:"managed_by_apps"`
			ManagedByArgoTunnel bool   `json:"managed_by_argo_tunnel"`
			Source              string `json:"source"`
		} `json:"meta"`
		CreatedOn  string `json:"created_on,omitempty"`
		ModifiedOn string `json:"modified_on,omitempty"`
	} `json:"result"`
}
type DNSRecord struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Prior   int    `json:"priority"`
	Ttl     int    `json:"ttl"`
	Type    string `json:"type"`
}

type Domain struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Id      string `json:"id"`
	ZoneId  string `json:"zone_id"`
}
type Zonedata struct {
	ZDomain string
	ZID     string
}
type Couldflare struct {
	ZonesId       string
	list          map[string][]*Domain
	Authorization string
	JsonDecode    T
	Zonelist      []Zonedata
}

func (this *Couldflare) req(method, id string, data io.Reader) io.ReadCloser {
	var url string
	if id != "getZones" {
		url = "https://api.cloudflare.com/client/v4/zones/" + this.ZonesId + "/dns_records/" + id
	} else {
		url = "https://api.cloudflare.com/client/v4/zones/"
	}
	get, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil
	}
	get.Header.Set("Content-Type", "application/json")
	get.Header.Set("Authorization", "Bearer "+this.Authorization)

	do, err := http.DefaultClient.Do(get)
	if err != nil {
		return nil
	}
	return do.Body
}

func (this *Couldflare) DeldomainName(domainName string) {
	this.GetAll()
	if v, ok := this.list[domainName]; ok {
		for _, v1 := range v {
			this.req("DELETE", v1.Id, nil)
		}
	}

}

func (this *Couldflare) DelId(id string) {
	this.req("DELETE", id, nil)
}
func (this *Couldflare) GetZones() string {
	res := this.req("GET", "getZones", nil)
	var zonesData Zone
	kk := json.NewDecoder(res)
	kk.Decode(&zonesData)
	marshal, err := json.Marshal(zonesData.Result)
	if err != nil {
		return ""
	}

	var tmp Zonedata
	for _, i2 := range zonesData.Result {
		tmp.ZID = i2.Id
		tmp.ZDomain = i2.Name
		this.Zonelist = append(this.Zonelist, tmp)
	}
	return string(marshal)

}
func (this *Couldflare) GetAll() map[string][]*Domain {

	this.list = make(map[string][]*Domain)

	data := this.req("GET", "", nil)
	kk := json.NewDecoder(data)
	for {
		err := kk.Decode(&this.JsonDecode)
		if err != nil {
			break
		}
	}
	for _, s := range this.JsonDecode.Result {

		this.list[s.Name] = append(this.list[s.Name], &Domain{Name: s.Name,
			Type:    s.Type,
			Id:      s.Id,
			Content: s.Content,
			ZoneId:  s.ZoneId,
		}) //fmt.Println(s.Name)
	}
	return this.list
}
func (this Couldflare) AddDNS(newDns DNSRecord) {

	payload, _ := json.Marshal(newDns)
	this.req("POST", "", bytes.NewReader(payload))

}
func (this Couldflare) ModifyDNSRecord(newDns DNSRecord, id string) {
	this.GetAll()
	_, ok := this.list[newDns.Name]
	if ok {
		payload, _ := json.Marshal(newDns)
		this.req("PUT", id, bytes.NewReader(payload))
	} else {
		this.AddDNS(newDns)
	}
}

func NewCouldFlareClient(ZonesId, Authorization string) *Couldflare {
	return &Couldflare{ZonesId: ZonesId,
		Authorization: Authorization,
	}
}
