package main

import (
    "flag"
    "log"

    "github.com/anerani/redyfi/dyficlient"
)

func main() {

    username := flag.String("username", "", "dy.fi username")
    password := flag.String("password", "", "dy.fi password")
    hostname := flag.String("hostname", "", "hostname to update")
    email := flag.String("mail", "", "email address for user agent header")
    flag.Parse()

    IPAddr := dyficlient.CheckIP()

    log.Printf("Seems like current IP address is: %s\n", IPAddr)
    log.Printf("Attempting to update...")

    responseBody, responseStatus := dyficlient.UpdateIP(username, password, hostname, email)

    log.Printf("Response status: %s\n", responseStatus)
    log.Printf("Response body: %s\n", responseBody)
}
