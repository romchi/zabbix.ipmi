package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type discoveryDevice struct {
	SensorID        string `json:"{#SENSOR_ID}"`
	SensorType      string `json:"{#SENSOR_TYPE}"`
	SensorName      string `json:"{#SENSOR_NAME}"`
	SensorUnit      string `json:"{#SENSOR_UNITS}"`
	SensorLowerCrit string `json:"{#SENSOR_LOWER_CRITICAL}"`
	SensorLowerWarn string `json:"{#SENSOR_LOWER_WARNING}"`
	SensorUpperCrit string `json:"{#SENSOR_UPPER_CRITICAL}"`
	SensorUpperWarn string `json:"{#SENSOR_UPPER_WARNING}"`
	SensorStatus    string `json:"{#SENSOR_STATUS}"`
}

var isLetter = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
var isLetterOrInt = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func main() {
	supportedSensors := []string{
		"Temperature",
		"Voltage",
		"Fan",
		"Physical_Security",
		"Power_Supply",
	}

	if len(os.Args) != 5 {
		fmt.Println("Usage:", os.Args[0], "HOST", "USERNAME", "PASSWORD", "SENSORS_TYPE")
		return
	}

	if !isIPv4(os.Args[1]) {
		fmt.Print("HOST - is not walid IPv4 address! \n")
		return
	}

	if !isLetter(os.Args[2]) {
		fmt.Print("USERNAME - not walid, accept only letters in username! \n")
		return
	}

	if !isLetterOrInt(os.Args[3]) {
		fmt.Print("PASSWORD - not walid, accept only letters and integers in password! \n")
		return
	}

	if !stringInSlice(os.Args[4], supportedSensors) {
		fmt.Printf("SENSORS_TYPE - not supported, accept next types: %v \n", strings.Join(supportedSensors, ","))
		return
	}

	runCommand(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
}

func runCommand(host, user, password, sensorType string) {
	binary, err := getBin("ipmi-sensors")
	if err != nil {
		fmt.Print("Unable find: ipmi-sensors! \n")
		return
	}

	args := []string{
		"--driver-type=LAN",
		fmt.Sprintf("--hostname=%v", host),
		fmt.Sprintf("--username=%v", user),
		fmt.Sprintf("--password=%v", password),
		"--privilege-level=USER",
		"--output-sensor-thresholds",
		"--ignore-not-available-sensors",
		"--comma-separated-output",
		"--no-header-output",
		"--quiet-cache",
		"--sdr-cache-recreate",
		fmt.Sprintf("--sensor-types=%v", sensorType),
	}
	cmd := exec.Command(binary, args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("ipmi-sensors connection error, try run command: \n%v %v\n", binary, strings.Join(args, " "))
		return
	}

	//fmt.Printf(string(out))
	// 0 - ID
	// 1 - Name
	// 2 - Type
	// 3 - Reading
	// 4 - Units
	// 5 - Lower NR
	// 6 - Lower C
	// 7 - Lower NC
	// 8 - Upper NC
	// 9 - Upper C
	// 10 - Upper NR
	// 11 - Event
	type discoveryDevice struct {
		SensorID   string `json:"{#SENSOR.ID}"`
		SensorName string `json:"{#SENSOR.NAME}"`
		//SensorType      string `json:"{#SENSOR.TYPE}"`
		SensorUnit      string `json:"{#SENSOR.UNITS}"`
		SensorLowerCrit string `json:"{#SENSOR.LOWER_CRIT}"`
		SensorLowerWarn string `json:"{#SENSOR.LOWER_WARN}"`
		SensorUpperCrit string `json:"{#SENSOR.UPPER_CRIT}"`
		SensorUpperWarn string `json:"{#SENSOR.UPPER_WARN}"`
		SensorStatus    string `json:"{#SENSOR.STATUS}"`
	}
	reader := csv.NewReader(bytes.NewReader(out))
	sensors := []discoveryDevice{}
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		sensors = append(sensors, discoveryDevice{
			SensorID:   line[0],
			SensorName: line[1],
			//SensorType:      line[2],
			SensorUnit:      line[4],
			SensorLowerCrit: line[6],
			SensorLowerWarn: line[7],
			SensorUpperCrit: line[9],
			SensorUpperWarn: line[10],
			SensorStatus:    line[11],
		})
	}
	s, _ := json.Marshal(sensors)

	fmt.Println(string(s))
}

func isIPv4(host string) bool {
	return net.ParseIP(host) != nil
}

func getBin(binFile string) (string, error) {
	location := []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/usr/local/bin", "/usr/local/sbin"}

	for _, path := range location {
		lookup := path + "/" + binFile
		fileInfo, err := os.Stat(path + "/" + binFile)
		if err != nil {
			continue
		}
		if !fileInfo.IsDir() {
			return lookup, nil
		}
	}
	return "", fmt.Errorf("Not found: '%v'", binFile)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
