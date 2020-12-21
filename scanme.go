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
	seentAt uint
	open    bool
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
	closedAfter := flag.Int("closed-after", 3, "port is considered closed only after missed X times")

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

	reverseIP := make(map[string]string)
	status := make(map[string]map[int]*openPort)
	// should probably make iter circular... but even it ticked every second, it's still 136 years to overflow...
	iter := uint(0)

	for {
		iter++
		// resolve hostnames for each iteration = up2date resolution
		// to update masscan argv with IP
		for targIndex, target := range flag.Args() {
			ip, err := net.LookupHost(target)
			realTarget := target
			if err == nil {
				realTarget = ip[0]
				params[targIndex+6] = ip[0]
			}
			reverseIP[realTarget] = target
			if _, ok := status[realTarget]; !ok {
				status[realTarget] = make(map[int]*openPort)
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
			for _, port := range result.Ports {
				if val, ok := status[result.IP][port.Port]; ok {
					val.seentAt = iter
				} else {
					// leave as "false" so it is marked as "new" later
					status[result.IP][port.Port] = &openPort{seentAt: iter, open: false}
				}
			}
		}
		for ip, ports := range status {
			var newPorts []int
			var closedPorts []int
			var totalPorts []int
			var output []string
			report := false

			for port, data := range ports {
				if !data.open && data.seentAt == iter {
					data.open = true
					newPorts = append(newPorts, port)
					totalPorts = append(totalPorts, port)
				} else if data.open {
					if iter-data.seentAt >= uint(*closedAfter) {
						closedPorts = append(closedPorts, port)
						data.open = false
					} else {
						totalPorts = append(totalPorts, port)
					}
				}
			}

			if len(newPorts) > 0 {
				output = append(output, fmt.Sprintf("*NEW*: `%s`", portsToString(newPorts)))
				report = true
			}
			if len(closedPorts) > 0 {
				output = append(output, fmt.Sprintf("*CLOSED*: `%s`", portsToString(closedPorts)))
				report = true
			}

			if report {
				output = append(output, fmt.Sprintf("*FINAL*: `%s`", portsToString(totalPorts)))
				log.Printf("%s (%s) - %s", reverseIP[ip], ip, strings.Join(output, " | "))
				if *notifyTokenPtr != "" {
					notify(*notifyTokenPtr, fmt.Sprintf("== %s (%s) ==\n%s", reverseIP[ip], ip, strings.Join(output, "\n")))
				}
			}
		}

		if *sleepIntervalPtr < 1 {
			break
		}
		time.Sleep(time.Duration(*sleepIntervalPtr) * time.Second)
	}
}
