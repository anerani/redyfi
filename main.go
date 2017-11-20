package main

import (
    "encoding/json"
    "flag"
    "log"
    "math/rand"
    "os"
    "os/user"
    "path/filepath"
    "reflect"
    "strings"
    "time"

    "github.com/anerani/redyfi/dyficlient"
)

type configs struct {
    Username string
    Password string
    Hostname string
    Email    string
}

type ipAddressState struct {
    IP        string
    Timestamp string
}

var configPathDefaults = []string{
    "Redyfi.json",
    "/etc/redyfi/Redyfi.json",
}

func readAndParseConfig(path string, config *configs) error {

    fileHandle, err := os.Open(path)
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

    usr, err := user.Current()

    if err != nil {
        log.Fatal(err)
    }

    configPathDefaults = append(configPathDefaults, filepath.Join(usr.HomeDir, ".redyfi", "Redyfi.json"))

    flag.String("username", "", "dy.fi username")
    flag.String("password", "", "dy.fi password")
    flag.String("hostname", "", "hostname to update")
    flag.String("email", "", "email address for user agent header")
    configPath := flag.String("configPath", "", "path to a configuration file")

    flag.Parse()

    config := &configs{}

    if *configPath != "" {
        if err := readAndParseConfig(*configPath, config); err != nil {
            log.Fatal(err)
        }
    } else {
        for _, path := range configPathDefaults {

            if _, err := os.Stat(path); os.IsNotExist(err) {
                continue
            }

            if err := readAndParseConfig(path, config); err != nil {
                log.Fatal(err)
            }
            break
        }
    }

    structReflection := reflect.ValueOf(config).Elem()

    // override config file settings with CLI arguments
    flag.VisitAll(func(f *flag.Flag) {
        value := f.Value.String()

        if value == "" {
            return
        }

        key := strings.Title(f.Name)
        field := structReflection.FieldByName(key)

        if field.IsValid() == false {
            return
        }

        field.SetString(value)
    })

    // all options are required (at least at the moment)
    // so check that all options have values

    structType := structReflection.Type()

    for i := 0; i < structReflection.NumField(); i++ {
        fieldInterface := structReflection.Field(i).Interface()

        if fieldInterface == reflect.Zero(reflect.TypeOf(fieldInterface)).Interface() {
            flag.Usage()
            log.Fatalf("Missing an argument for: %s", structType.Field(i).Name)
        }
    }

    IPAddr := dyficlient.CheckIP()

    log.Printf("Seems like current IP address is: %s\n", IPAddr)
    log.Printf("Attempting to perform an initial update...")

    responseBody, responseStatus := dyficlient.UpdateIP(config.Username, config.Password, config.Hostname, config.Email)

    log.Printf("Response status: %s\n", responseStatus)
    log.Printf("Response body: %s\n", responseBody)

    // dy.fi spesification recommends using slightly random weekly interval
    // for updates to avoid congestion
    oneWeekFromNowInMinutes := time.Duration((60*24*6 + rand.Intn(60*23+59))) * time.Minute

    c := time.Tick(oneWeekFromNowInMinutes)

    for _ = range c {
        log.Print("About one week has passed. Attempting an update...")
        IPAddr := dyficlient.CheckIP()

        log.Printf("Seems like current IP address is: %s\n", IPAddr)

        responseBody, responseStatus := dyficlient.UpdateIP(config.Username, config.Password, config.Hostname, config.Email)

        log.Printf("Response status: %s\n", responseStatus)
        log.Printf("Response body: %s\n", responseBody)
        log.Print("Going back to sleep.")
    }

}
