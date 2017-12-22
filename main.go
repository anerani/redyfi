package main

import (
	"bytes"
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

	"github.com/anerani/redyfi/dyfi"
)

var configPathDefaults = []string{
	"Redyfi.json",
	"/etc/redyfi/Redyfi.json",
}

type configs struct {
	Username string
	Password string
	Hostname string
	Email    string
}

func readAndDecodeConfig(path string, config *configs) error {

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

func validateConfig(config *configs) {
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

	structType := structReflection.Type()

	for i := 0; i < structReflection.NumField(); i++ {
		fieldInterface := structReflection.Field(i).Interface()

		if fieldInterface == reflect.Zero(reflect.TypeOf(fieldInterface)).Interface() {
			flag.Usage()
			log.Fatalf("[ERROR] Missing an argument for: %s", structType.Field(i).Name)
		}
	}
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

	config := &configs{}

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
	validateConfig(config)

	dyfiClient := &dyfi.Client{
		Username: config.Username,
		Password: config.Password,
		Hostname: config.Hostname,
		Email:    config.Email,
	}

	IPAddr, err := dyfiClient.CheckIP()

	log.Printf("[INFO] Seems like current IP address is: %s\n", IPAddr)
	log.Printf("[INFO] Attempting to perform an initial update...")

	err = dyfiClient.UpdateIP()
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
			HourlyIPAddrCheck, err := dyfiClient.CheckIP()

			if err != nil {
				log.Print("[ERROR]: Checking IP address failed.")
				log.Print(err)
				return
			}

			log.Printf("[INFO] Seems like current IP address is: %s\n", IPAddr)

			if bytes.Equal(IPAddr, HourlyIPAddrCheck) == false {
				log.Print("[INFO] Address has changed since last update. Updating before weekly update...")
				IPAddr = HourlyIPAddrCheck

				err := dyfiClient.UpdateIP()
				if err != nil {
					log.Fatal(err)
				}

				weeklyTick = time.NewTicker(oneWeekDuration)
			} else {
				log.Println("[INFO] No need to update.")
			}

		case <-weeklyTick.C:
			log.Print("[INFO] About one week has passed. Attempting an update...")

			err = dyfiClient.UpdateIP()
			if err != nil {
				log.Print(err)
			}

			log.Print("[INFO] Going back to sleep.")
		}
	}

}
