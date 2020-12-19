package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type result struct {
	IP        string `json:"ip"`
	Timestamp string `json:"timestamp"`
	Ports     []port `json:"ports"`
}

type port struct {
	Port   int    `json:"port"`
	Proto  string `json:"proto"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	TTL    int    `json:"ttl"`
}

type openPort struct {
	SeentAt int
	Open    bool
}

// https://siongui.github.io/2018/03/14/go-set-difference-of-two-arrays/
func difference(a, b []int) (diff []int) {
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

func portsToString(ports []int) string {
	r := make([]string, len(ports))
	for pi, port := range ports {
		r[pi] = strconv.Itoa(port)
	}
	return strings.Join(r, ",")
}

func notify(token string, message string) bool {
	resp, err := http.PostForm(
		"https://tgbots.skmobi.com/pushit/"+token,
		url.Values{"msg": {message}, "format": {"Markdown"}},
	)
	if err != nil {
		log.Println(resp)
		log.Println(err)
		return false
	}
	return true
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

	var results []result
	params := append(
		[]string{
			"-oJ", file.Name(),
			"--rate", *rateLimitPtr,
			"-p", "1-65535",
		},
		flag.Args()...,
	)

	var lastRun map[string][]int

	for {
		summary := make(map[string][]int)
		// resolve hostnames for each iteration = up2date resolution
		for targIndex, target := range flag.Args() {
			ip, err := net.LookupHost(target)
			if err == nil {
				summary[ip[0]] = []int{}
				params[targIndex+6] = ip[0]
			} else {
				summary[target] = []int{}
			}
		}
		cmd := exec.Command(*masscanPathPtr, params...)
		log.Printf("Scanning %v", flag.Args())
		if *showOutputPtr {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		start := time.Now()
		err = cmd.Run()
		if err != nil {
			log.Printf("Scan error: %v", err)
		}
		log.Printf("Scan finished in %v", time.Since(start))

		jsonFile, err := os.Open(file.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &results)
		if err != nil {
			results = []result{}
		}

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
			newPorts := difference(v, lastRun[k])
			closedPorts := difference(lastRun[k], v)
			report := false
			var output []string

			if len(newPorts) > 0 {
				output = append(output, fmt.Sprintf("*NEW*: `%s`", portsToString(newPorts)))
				report = true
			}
			if len(closedPorts) > 0 {
				output = append(output, fmt.Sprintf("*CLOSED*: `%s`", portsToString(closedPorts)))
				report = true
			}

			if report {
				output = append(output, fmt.Sprintf("*FINAL*: `%s`", portsToString(v)))
				log.Printf("%s - %s", k, strings.Join(output, " | "))
				if *notifyTokenPtr != "" {
					notify(*notifyTokenPtr, fmt.Sprintf("== %s ==\n%s", k, strings.Join(output, "\n")))
				}

			}

		}
		lastRun = summary

		if *sleepIntervalPtr < 1 {
			break
		}
		time.Sleep(time.Duration(*sleepIntervalPtr) * time.Second)
	}
}
