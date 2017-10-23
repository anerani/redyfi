package main

import (
    "encoding/json"
    "flag"
    "log"
    "os"

    "github.com/anerani/redyfi/dyficlient"
)

type configs struct {
    Username string
    Password string
    Hostname string
    Email    string
}

var configPathDefaults = [...]string{
    "Redyfi.json",
    "/etc/redyfi/Redyfi.json",
}

func readAndParseConfig(path *string, config *configs) error {
    fileHandle, err := os.Open(*path)
    if err != nil {
        return err
    }

    jsonDecoder := json.NewDecoder(fileHandle)

    err = jsonDecoder.Decode(&config)

    if err != nil {
        return err
    }
    return nil
}

func main() {

    username := flag.String("username", "", "dy.fi username")
    password := flag.String("password", "", "dy.fi password")
    hostname := flag.String("hostname", "", "hostname to update")
    email := flag.String("mail", "", "email address for user agent header")
    configPath := flag.String("configPath", "", "path to a configuration file")

    flag.Parse()

    config := &configs{}

    if *configPath != "" {
        if err := readAndParseConfig(configPath, config); err != nil {
            log.Fatal(err)
        }
    } else {
        for _, path := range configPathDefaults {

            if _, err := os.Stat(path); os.IsNotExist(err) {
                continue
            }
            if err := readAndParseConfig(configPath, config); err != nil {
                log.Fatal(err)
            }
            break
        }
    }

    if *username != "" {
        config.Username = *username
    }
    if *password != "" {
        config.Password = *password
    }
    if *hostname != "" {
        config.Hostname = *hostname
    }
    if *email != "" {
        config.Email = *email
    }

    IPAddr := dyficlient.CheckIP()

    log.Printf("Seems like current IP address is: %s\n", IPAddr)
    log.Printf("Attempting to update...")

    responseBody, responseStatus := dyficlient.UpdateIP(config.Username, config.Password, config.Hostname, config.Email)

    log.Printf("Response status: %s\n", responseStatus)
    log.Printf("Response body: %s\n", responseBody)
}
