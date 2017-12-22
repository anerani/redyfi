package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/anerani/redyfi/dyfi"
)

var configPathDefaults = []string{
	"Redyfi.json",
	"/etc/redyfi/Redyfi.json",
}

func readAndDecodeConfig(path string, config *dyfi.ClientConfig) error {

	fileHandle, err := os.Open(path)
	defer fileHandle.Close()

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

func validateConfig(config *dyfi.ClientConfig) error {
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

	// check that all configuration parameters have values

	var configMap map[string]interface{}
	tmp, err := json.Marshal(config)

	if err != nil {
		return err
	}

	json.Unmarshal(tmp, &configMap)

	for key, value := range configMap {
		if value == "" {
			return fmt.Errorf("[ERROR] Missing a value for: %s", key)
		}
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
	runAsDaemon := flag.Bool("daemon", false, "run redyfi as a service")
	configPath := flag.String("configPath", "", "path to a configuration file")

	flag.Parse()

	config := &dyfi.ClientConfig{}

	if *configPath != "" {
		if err := readAndDecodeConfig(*configPath, config); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, path := range configPathDefaults {

			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}

			if err := readAndDecodeConfig(path, config); err != nil {
				log.Fatal(err)
			}
			break
		}
	}
	if err := validateConfig(config); err != nil {
		log.Fatal(err)
	}

	client := dyfi.NewClient(config)

	IPAddr, err := client.CheckIP()

	log.Printf("[INFO] Seems like current IP address is: %s\n", IPAddr)
	log.Printf("[INFO] Attempting to perform an initial update...")

	err = client.UpdateIP()
	if err != nil {
		log.Fatal(err)
	}

	if *runAsDaemon == false {
		return
	}

	log.Println("[Info] Going to sleep.")

	// dy.fi spesification recommends using slightly random weekly interval
	// for updates to avoid congestions (https://www.dy.fi/page/specification)
	oneWeekDuration := time.Duration((60*24*6 + rand.Intn(60*23+59))) * time.Minute
	oneHourDuration := 60 * time.Minute
	weeklyTick := time.NewTicker(oneWeekDuration)
	hourlyTick := time.NewTicker(oneHourDuration)

	for {
		select {

		case <-hourlyTick.C:
			HourlyIPAddrCheck, err := client.CheckIP()

			if err != nil {
				log.Println("[ERROR]: Checking IP address failed.")
				log.Print(err)
				return
			}

			log.Printf("[INFO] Seems like current IP address is: %s\n", IPAddr)

			if bytes.Equal(IPAddr, HourlyIPAddrCheck) == false {
				log.Println("[INFO] Address has changed since last update. Updating before weekly update...")
				IPAddr = HourlyIPAddrCheck

				err := client.UpdateIP()
				if err != nil {
					log.Fatal(err)
				}

				weeklyTick = time.NewTicker(oneWeekDuration)
			} else {
				log.Println("[INFO] No need to update.")
			}

		case <-weeklyTick.C:
			log.Println("[INFO] About one week has passed. Attempting an update...")

			err = client.UpdateIP()
			if err != nil {
				log.Print(err)
			}

			log.Print("[INFO] Going back to sleep.")
		}
	}

}
