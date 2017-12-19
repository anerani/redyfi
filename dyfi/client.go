package dyfi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

var (
	checkIPURL      = "http://checkip.dy.fi/"
	updateIPBaseURL = "https://www.dy.fi/nic/update?hostname="
)

type Client struct {
	Username string
	Password string
	Hostname string
	Email    string
}

// CheckIP gets the current IP address using the dy.fi service.
func (*Client) CheckIP() ([]byte, error) {
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
	updateIPURL := updateIPBaseURL + c.Hostname

	request, err := http.NewRequest("GET", updateIPURL, nil)
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.Username, c.Password)
	request.Header.Set("User-Agent", "redyfi/0.2.0 ("+c.Email+")")

	httpClient := &http.Client{}

	log.Println("[INFO] Updating...")
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("[ERROR] Requesting IP update failed. Server returned: %s (%s)", body, response.Status)
	}

	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Println("[INFO] Update successful.")

	return nil
}
