package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type UpdateCreateDNSResult struct {
	Result struct {
		ID        string `json:"id"`
		ZoneID    string `json:"zone_id"`
		ZoneName  string `json:"zone_name"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Content   string `json:"content"`
		Proxiable bool   `json:"proxiable"`
		Proxied   bool   `json:"proxied"`
		TTL       int    `json:"ttl"`
		Locked    bool   `json:"locked"`
		Meta      struct {
			AutoAdded           bool   `json:"auto_added"`
			ManagedByApps       bool   `json:"managed_by_apps"`
			ManagedByArgoTunnel bool   `json:"managed_by_argo_tunnel"`
			Source              string `json:"source"`
		} `json:"meta"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

// HTTp数据包
type HttpData struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}
type CloudflareAPI struct {
	ZoneID     string
	Host       string
	APIToken   string
	BaseURL    string
	httpClient *http.Client
}

type Record struct {
	ID      string     `json:"id"`
	Type    RecordType `json:"type"`
	Content string     `json:"content"`
	Name    string     `json:"name"`
	Proxied bool       `json:"proxied"`
}

type RecordType = string

const (
	RecordTypeA    = RecordType("A")
	RecordTypeAAAA = RecordType("AAAA")
)

type RecordResponse struct {
	Result []Record `json:"result"`
}

func NewCloudflareClient(token string, zoneID string, host string) (*CloudflareAPI, error) {
	api := CloudflareAPI{
		ZoneID:   zoneID,
		Host:     host,
		APIToken: token,
		BaseURL:  "https://api.cloudflare.com/client/v4",
	}

	if api.httpClient == nil {
		api.httpClient = http.DefaultClient
	}

	return &api, nil
}

func (api *CloudflareAPI) ListDNSRecords(recType RecordType) ([]Record, error) {
	uri := fmt.Sprintf("/zones/%s/dns_records?type=%s&name=%s", api.ZoneID, recType, api.Host)
	resp, err := api.request("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	var r *RecordResponse
	err = json.Unmarshal(resp, &r)

	if err != nil {
		return nil, err
	}

	return r.Result, nil
}

func (api *CloudflareAPI) UpdateDNSRecord(record Record) error {
	uri := fmt.Sprintf("/zones/%s/dns_records/%s", api.ZoneID, record.ID)

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(record)

	_, err := api.request("PUT", uri, payload)
	if err != nil {
		return err
	}

	return nil
}

func (api *CloudflareAPI) request(method string, uri string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, api.BaseURL+uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIToken))

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Status code not 200 but %v, body: %v", err, string(respBody))
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
func Del(Authorization, ZonesId, id string) {
	var url string
	if id != "getZones" {
		url = "https://api.cloudflare.com/client/v4/zones/" + ZonesId + "/dns_records/" + id
	} else {
		url = "https://api.cloudflare.com/client/v4/zones/"
	}
	get, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	get.Header.Set("Content-Type", "application/json")
	get.Header.Set("Authorization", "Bearer "+Authorization)

	do, err := http.DefaultClient.Do(get)
	if err != nil {
		return
	}
	fmt.Println(do.Status)
}
func CreateDNSRecord(IpProtocol string, IpAddress string, IpDpmain string, Authorization, ZonesId string) UpdateCreateDNSResult {
	var createDNSResult UpdateCreateDNSResult

	createData := &HttpData{
		Name:    IpDpmain,
		Content: IpAddress,
		TTL:     100,
		Proxied: false,
	}
	createData.Type = IpProtocol
	createDataJson, err := json.Marshal(createData)
	if err != nil {
		log.Println(err)
	}
	client := &http.Client{}
	rootUrl := "https://api.cloudflare.com/client/v4/zones/" + ZonesId + "/dns_records/"
	req, _ := http.NewRequest("POST", rootUrl, bytes.NewReader(createDataJson))
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+Authorization)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return createDNSResult
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &createDNSResult)
	if err != nil {
		panic(err)
	}
	return createDNSResult
}
