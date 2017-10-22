package dyficlient

import (
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

// CheckIP gets the current IP address using the dy.fi service.
func CheckIP() []byte {
    response, err := http.Get(checkIPURL)

    if err != nil {
        log.Fatal(err)
    }

    defer response.Body.Close()

    body, err := ioutil.ReadAll(response.Body)

    if err != nil {
        log.Fatal(err)
    }

    // catch only the IP address from the result string. response body format:
    // Current IP Address: [0-255].[0-255].[0-255].[0-255]
    IPMatcher, err := regexp.Compile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)

    if err != nil {
        log.Fatal(err)
    }

    IPAddr := IPMatcher.Find(body)
    if IPAddr == nil {
        log.Fatal(err)
    }

    // validate the IP address
    if net.ParseIP(string(IPAddr)) == nil {
        log.Fatal(err)
    }

    return IPAddr
}

// UpdateIP sends a refresh request to the dy.fi server to update current IP address
// pointing to a hostname. Returns the response body and status as a string.
func UpdateIP(username *string, password *string, hostname *string, email *string) ([]byte, string) {
    updateIPURL := updateIPBaseURL + *hostname

    log.Printf(updateIPURL + "\n")
    request, err := http.NewRequest("GET", updateIPURL, nil)

    if err != nil {
        log.Fatal(err)
    }

    request.SetBasicAuth(*username, *password)
    request.Header.Set("User-Agent", "redyfi/0.0.2 ("+*email+")")

    client := &http.Client{}

    response, err := client.Do(request)

    if err != nil {
        log.Fatal(err)
    }

    defer response.Body.Close()

    body, err := ioutil.ReadAll(response.Body)

    if err != nil {
        log.Fatal(err)
    }

    return body, response.Status
}
