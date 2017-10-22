package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var (
	checkIPURL      = "http://checkip.dy.fi/"
	updateIPBaseURL = "https://www.dy.fi/nic/update?hostname="
)

func main() {

	userName := flag.String("username", "", "dy.fi username")
	passWord := flag.String("password", "", "dy.fi password")
	hostName := flag.String("hostname", "", "hostname to update")
	eMail := flag.String("mail", "", "email address for user agent header")
	flag.Parse()

	response, err := http.Get(checkIPURL)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	IPMatcher, err := regexp.Compile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)

	if err != nil {
		log.Fatal(err)
	}

	IPAddr := IPMatcher.Find(body)

	if IPAddr == nil {
		log.Fatal(err)

	}
	log.Printf("Seems like current IP address is: %s\n", IPAddr)
	log.Printf("Attempting to update...")

	updateIPURL := updateIPBaseURL + *hostName

	log.Printf(updateIPURL + "\n")
	request, err := http.NewRequest("GET", updateIPURL, nil)

	if err != nil {
		log.Fatal(err)
	}

	request.SetBasicAuth(*userName, *passWord)
	request.Header.Set("User-Agent", "redyfi/0.0.1 ("+*eMail+")")

	client := &http.Client{}

	response, err = client.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response status: %s\n", response.Status)
	log.Printf("Response body: %s\n", body)
}
