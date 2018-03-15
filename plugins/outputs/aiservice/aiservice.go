package aiservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai/event"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// Aiservice struct is the primary data structure for the plugin
type Aiservice struct {
	LoginURL               string `toml:"login_url"`
	URL                    string `toml:"url"`
	Username               string
	Password               string
	authority              string
	refreshTime            time.Time
	authorityClient        *http.Client
	aiserviceDbrelayClient *http.Client
}
type reqType int

const (
	authority = reqType(0)
	metric    = reqType(1)
)

var (
	maxRequestTimes        = 3
	loginInterval          = time.Minute * 50
	loginURL               = "https://api.aiservice.io/devdcaccount/v1/login"
	aiserviceURL           = "https://api.aiservice.io/devdp/v1/metrics/"
	writeAmountOfBatchData = 100
)

var sampleConfig = `
# Login username and password
username = ""	#required
password = ""	#required
`

// Description uppon outputs.aiservice
func (i *Aiservice) Description() string {
	return "Diskprophet AIService output"
}

// SampleConfig would return sampleConfig
func (i *Aiservice) SampleConfig() string {
	return sampleConfig
}

func checkURL(rawurl string) error {
	// parse URL:
	u, err := url.Parse(rawurl)
	if err != nil {
		return fmt.Errorf("error parsing config.URL: %s", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("config.URL scheme must be http(s), got %s", u.Scheme)
	}
	return nil
}

// Connect would login aiservice
// and get authority
func (i *Aiservice) Connect() error {
	if i.Username == "" || i.Password == "" {
		log.Printf("E! Need to put username and password informations for aiservice configuration")
		return nil
	}

	// if login_url given in conf
	if i.LoginURL != "" {
		loginURL = i.LoginURL
		err := checkURL(loginURL)
		if err != nil {
			return err
		}
	}

	data := map[string]string{}

	data["email"] = i.Username
	data["password"] = i.Password

	payloadBytes, _ := json.Marshal(data)
	body := bytes.NewReader(payloadBytes)

	err := i.getAuthority(loginURL, body)
	if err != nil {
		return err
	}

	return nil
}

// Close connection
func (i *Aiservice) Close() error {
	// Close connection to the URL here
	return nil
}

func (i *Aiservice) Write(metrics []telegraf.Metric) error {
	if i.authority == "" {
		log.Println("I! Need to login again to get authority!")
		i.Connect()
		return nil
	}
	// if url given in conf
	if i.URL != "" {
		aiserviceURL = i.URL
		err := checkURL(aiserviceURL)
		if err != nil {
			return err
		}
	}
	// check authority
	i.checkAuthoriyExpiration()

	// add heartBeat interval event log
	if err := event.AddIntervalAgentHeartbeatMetric(&metrics); err != nil {
		return err
	}

	// block metrics
	measurementsOfBlocksMetricsArray := createBlockMetricsArray(metrics)

	for measurement, blocksMetricsArray := range measurementsOfBlocksMetricsArray {

		for _, blockMetrics := range blocksMetricsArray {
			payloadBytes := buildPayloadOfAiServiceCloudAPIMetrics(blockMetrics)
			body := bytes.NewReader(payloadBytes)
			_, err := i.makeAndDoRequest(metric, aiserviceURL+measurement, body)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func newAiservice() *Aiservice {
	return &Aiservice{
		authorityClient:        &http.Client{Timeout: time.Second * 10},
		aiserviceDbrelayClient: &http.Client{Timeout: time.Second * 30},
	}
}

func init() {
	outputs.Add("aiservice", func() telegraf.Output { return newAiservice() })
}

func retryCrit(resp *http.Response, err error) bool {
	if err != nil {
		// encount err with "INTERNAL_ERROR" or "Client.Timeout" would return true
		return (strings.ContainsAny(fmt.Sprintf("%s", err), "INTERNAL_ERROR & Client.Timeout"))
	}
	return (resp.StatusCode == http.StatusBadGateway)
}

func (i *Aiservice) makeAndDoRequest(reqType reqType, url string, body io.Reader) (*http.Response, error) {
	req, err := i.makeRequest(reqType, url, body)
	if err != nil {
		return nil, fmt.Errorf("fail to make request with request type %v\n ", reqType)
	}

	for n := 1; n <= maxRequestTimes; n++ {
		resp, err := i.doRequest(reqType, req)
		if retryCrit(resp, err) {
			log.Printf("W! %s Retrying request %d ", err, n)
			continue
		}
		return resp, err
	}

	return nil, fmt.Errorf("out of maximum request times : %d ", maxRequestTimes)
}

func (i *Aiservice) doRequest(reqType reqType, req *http.Request) (*http.Response, error) {
	switch reqType {
	case authority:
		resp, err := i.authorityClient.Do(req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	case metric:
		resp, err := i.aiserviceDbrelayClient.Do(req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	}
	return nil, fmt.Errorf("No such request type")
}

func (i *Aiservice) makeRequest(reqType reqType, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("can make an new request %v\n ", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if reqType == metric {
		req.Header.Set("Authorization", i.authority)
	}

	return req, nil
}

// set value for i.Authority
func (i *Aiservice) getAuthority(url string, body io.Reader) error {
	i.refreshTime = time.Now()

	resp, err := i.makeAndDoRequest(authority, url, body)
	if err != nil {
		return fmt.Errorf("Unauthority, %v\n ", err)
	}

	var f map[string]string
	jsonResponse := f
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &jsonResponse)

	i.authority = jsonResponse["Authentication"]

	defer resp.Body.Close()

	return nil
}

// merge metric.Fields and metric.Tags
// return payloadbytes
func buildPayloadOfAiServiceCloudAPIMetrics(metrics []telegraf.Metric) []byte {
	var payloadBytes []byte

	type Tmp struct {
		Points []map[string]interface{} `json:"points"`
	}

	var payload = Tmp{Points: []map[string]interface{}{}}

	// claim a null map to restore maps
	var data map[string]interface{}

	for _, metric := range metrics {
		data = map[string]interface{}{}

		data["time"] = fmt.Sprintf("%d", metric.Time().UnixNano())

		// merge both mertic.Fields() and metric.Tags()
		for k, v := range metric.Tags() {
			data[k] = v
		}
		for k, v := range metric.Fields() {
			data[k] = v
		}

		payload.Points = append(payload.Points, data)
	}

	payloadBytes, _ = json.Marshal(payload)

	return payloadBytes

}

func (i *Aiservice) checkAuthoriyExpiration() error {
	if (time.Now().UnixNano() - i.refreshTime.UnixNano()) > loginInterval.Nanoseconds() {
		err := i.Connect()
		if err != nil {
			return err
		}
	}
	return nil
}

// make block before send request
// avoid send request frequency
func createBlockMetricsArray(metrics []telegraf.Metric) map[string][][]telegraf.Metric {
	measurementsOfBlocksMetricsArray := map[string][][]telegraf.Metric{}
	for _, metric := range metrics {

		measurement := metric.Name()
		blocksMetricArray, exist := measurementsOfBlocksMetricsArray[measurement]

		if exist == true {

			lastBlockMetricsIndex := len(blocksMetricArray) - 1
			lastBlockMetrics := blocksMetricArray[lastBlockMetricsIndex]

			if len(lastBlockMetrics) < writeAmountOfBatchData {

				lastBlockMetrics = append(lastBlockMetrics, metric)
				measurementsOfBlocksMetricsArray[measurement][lastBlockMetricsIndex] = lastBlockMetrics
			} else {

				blockMetrics := []telegraf.Metric{metric}
				measurementsOfBlocksMetricsArray[measurement] = append(measurementsOfBlocksMetricsArray[measurement], blockMetrics)
			}

		} else {
			blockMetrics := []telegraf.Metric{metric}
			measurementsOfBlocksMetricsArray[measurement] = append(measurementsOfBlocksMetricsArray[measurement], blockMetrics)
		}
	}
	return measurementsOfBlocksMetricsArray
}
