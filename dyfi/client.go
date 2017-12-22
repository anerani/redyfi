package dyfi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var (
	checkIPURL         = "http://checkip.dy.fi/"
	updateIPBaseURL    = "https://www.dy.fi/nic/update?hostname="
	bodyStatusMessages = map[string]string{
		"nohost": "No 'hostname' CGI parameter given in the request, or the hostname is not allocated for the user.",
		"nofqdn": "The given hostname is not a valid .dy.fi FQDN.",
		"badip":  "The client IP address is not a valid IP address, or is not registered to a Finnish organisation.",
		"dnserr": "The request failed due to a technical problem at the dy.fi service.",
		"abuse":  "The request was denied because of abuse (too many requests in a short time).",
		"nochg":  "The request was valid and processed, but did not cause a change in the DNS information since the information had not changed since last update (the client IP address had not changed).",
		"good":   "The request was valid and processed successfully, and caused the hostname to be pointed to the IP address returned.",
	}
)

type ClientConfig struct {
	Username string
	Password string
	Hostname string
	Email    string
}

type Client struct {
	Client   *http.Client
	Settings ClientConfig
}

func NewClient(config *ClientConfig) *Client {

	transport := &http.Transport{
		DisableKeepAlives: true,
	}

	return &Client{
		Client: &http.Client{
			Transport: transport,
		},
		Settings: *config,
	}
}

// CheckIP gets the current IP address using the dy.fi service.
func (c *Client) CheckIP() ([]byte, error) {
	response, err := http.Get(checkIPURL)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// catch only the IP address from the result string. response body format:
	// Current IP Address: [0-255].[0-255].[0-255].[0-255]
	IPMatcher, err := regexp.Compile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	if err != nil {
		return nil, err
	}

	IPAddr := IPMatcher.Find(body)
	if IPAddr == nil || net.ParseIP(string(IPAddr)) == nil {
		return nil, fmt.Errorf("No valid IP address found. Server response was: %s", body)
	}

	return IPAddr, nil
}

// UpdateIP sends a refresh request to the dy.fi server to update current IP address
// pointing to a hostname. Returns the response body and status as a string.
func (c *Client) UpdateIP() error {
	updateIPURL := updateIPBaseURL + c.Settings.Hostname

	request, err := http.NewRequest("GET", updateIPURL, nil)
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.Settings.Username, c.Settings.Password)
	request.Header.Set("User-Agent", "redyfi/1.0.3 ("+c.Settings.Email+")")

	log.Println("[INFO] Updating...")
	response, err := c.Client.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return err
	}
	bodyString := strings.TrimSpace(string(body))

	if response.StatusCode != 200 {
		return fmt.Errorf("[ERROR] Requesting IP update failed. Server returned: %s (%s)", bodyString, response.Status)
	}

	if _, exists := bodyStatusMessages[bodyString]; exists == false {
		return fmt.Errorf("[ERROR] Unknown status message returned by the server: %s (%s)", bodyString, response.Status)
	}

	statusMessage := bodyStatusMessages[bodyString]

	if bodyString != "good" && bodyString != "nochg" {
		return fmt.Errorf("[ERROR] %s", statusMessage)
	}
	log.Println("[INFO]", statusMessage)
	log.Println("[INFO] Update successful.")

	return nil
}
