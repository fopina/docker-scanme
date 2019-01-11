package main

import (
	"log"
	"flag"
	"os/exec"
	"io/ioutil"
	"os"
	"time"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"strconv"
	"net/http"
	"net/url"
)

type Result struct {
	IP   		string `json:"ip"`
	Timestamp   string `json:"timestamp"`
	Ports    	[]Port `json:"ports"`
}

type Port struct {
	Port   	int    `json:"port"`
	Proto   string `json:"proto"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	TTL   	int    `json:"ttl"`
}

// https://siongui.github.io/2018/03/14/go-set-difference-of-two-arrays/
func Difference(a, b []int) (diff []int) {
      m := make(map[int]bool)

      for _, item := range b {
              m[item] = true
      }

      for _, item := range a {
              if _, ok := m[item]; !ok {
                      diff = append(diff, item)
              }
      }
      return
}

func PortsToString(ports []int) string {
	r := make([]string, len(ports))
	for pi, port := range ports {
		r[pi] = strconv.Itoa(port)
	}
	return strings.Join(r, ",")
}

func Notify(token string, message string) bool {
	_, err := http.PostForm(
		"https://tgbots.skmobi.com/pushit/" + token,
		url.Values{"msg": {message}, "format": {"Markdown"}},
	)
	return err == nil
}

func main() {
	masscanPathPtr := flag.String("path", "/usr/bin/masscan", "path to masscan binary")
	sleepIntervalPtr := flag.Int("sleep", 1800, "number of seconds to sleep between re-scans, set to 0 to disable")
	rateLimitPtr := flag.String("rate", "100", "masscan rate")
	showOutputPtr := flag.Bool("show", false, "show masscan output")
	notifyTokenPtr := flag.String("token", "", "PushItBot token for notifications")

	flag.Usage = func() {
        fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] target [target ...]\n", os.Args[0])
        fmt.Fprintf(flag.CommandLine.Output(), "\nContinuously scan one (or more) targets\n")
        fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
        flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 { 
		log.Fatal("specify at least one target (IP or hostname): check -h")
	}

	file, err := ioutil.TempFile("", "scanme.*.json")
	if err != nil {
	    log.Fatal(err)
	}
	defer os.Remove(file.Name())

	var results []Result
	params := append(
		[]string{
			"-oJ", file.Name(),
			"-rate", *rateLimitPtr,
		},
		flag.Args()...
	)

	var lastRun map[string][]int

	for {
		summary := make(map[string][]int)
		// resolve hostnames for each iteration = up2date resolution
		for targ_index, target := range flag.Args() {
			ip, err := net.LookupHost(target)
			if err == nil {
				params[targ_index + 2] = ip[0]
			}
		}
		cmd := exec.Command(*masscanPathPtr, params...)
		log.Printf("Scanning %v", flag.Args())
		if (*showOutputPtr) {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		start := time.Now()
		err = cmd.Run()
		if (err != nil) {
			log.Printf("Scan error: %v", err)
		}
		log.Printf("Scan finished in %v", time.Since(start))

		jsonFile, err := os.Open(file.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		json.Unmarshal(byteValue, &results)

		for _, result := range results {
			ipArray, exists := summary[result.IP]
			if !exists {
				ipArray = []int{}
			}
			for _, port := range result.Ports {
				summary[result.IP] = append(ipArray, port.Port)
			}
		}
		
		for k, v := range summary {
			newPorts := Difference(v, lastRun[k])
			closedPorts := Difference(lastRun[k], v)
			report := false
			var output []string

			if len(newPorts) > 0 {
				output = append(output, fmt.Sprintf("*NEW*: `%s`", PortsToString(newPorts)))
				report = true
			}
			if len(closedPorts) > 0 {
				output = append(output, fmt.Sprintf("*CLOSED*: `%s`", PortsToString(closedPorts)))
				report = true
			}

			if report {
				output = append(output, fmt.Sprintf("*FINAL*: `%s`", PortsToString(v)))
				log.Printf("%s - %s", k, strings.Join(output, " | "))
				if *notifyTokenPtr != "" {
					Notify(*notifyTokenPtr, fmt.Sprintf("== %s ==\n%s", k, strings.Join(output,"\n")))
				}

			}

		}
		lastRun = summary

		if (*sleepIntervalPtr < 1) {
			break
		}
		time.Sleep(time.Duration(*sleepIntervalPtr) * time.Second)
	}
}
