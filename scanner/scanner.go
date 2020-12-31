package scanner

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

// Scanner ...
type Scanner struct {
	MasscanPath string
	RateLimit   string
	ShowOutput  bool
	NotifyToken string
	ClosedAfter uint
	Targets     []string
	tempFile    string
	scanArgs    []string
	reverseIP   map[string]string
	status      map[string]map[int]*openPort
	iter        uint
}

func portsToString(ports []int) string {
	r := make([]string, len(ports))
	for pi, port := range ports {
		r[pi] = strconv.Itoa(port)
	}
	return strings.Join(r, ",")
}

func (s *Scanner) notify(message string) bool {
	if s.NotifyToken == "" {
		return false
	}
	resp, err := http.PostForm(
		"https://tgbots.skmobi.com/pushit/"+s.NotifyToken,
		url.Values{"msg": {message}, "format": {"Markdown"}},
	)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (s *Scanner) runScanner() {
	// resolve hostnames for each iteration = up2date resolution
	// to update masscan argv with IP
	for targIndex, target := range flag.Args() {
		ip, err := net.LookupHost(target)
		realTarget := target
		if err == nil {
			realTarget = ip[0]
			s.scanArgs[targIndex+6] = ip[0]
		}
		s.reverseIP[realTarget] = target
		if _, ok := s.status[realTarget]; !ok {
			s.status[realTarget] = make(map[int]*openPort)
		}
	}
	cmd := exec.Command(s.MasscanPath, s.scanArgs...)
	log.Printf("Scanning %v", flag.Args())
	if s.ShowOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	start := time.Now()
	err := cmd.Run()
	if err != nil {
		log.Printf("Scan error: %v", err)
	}
	log.Printf("Scan finished in %v", time.Since(start))
}

func (s *Scanner) parseResults() error {
	var results []result
	jsonFile, err := os.Open(s.tempFile)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &results)
	if err != nil {
		return err
	}

	for _, result := range results {
		for _, port := range result.Ports {
			if val, ok := s.status[result.IP][port.Port]; ok {
				val.seentAt = s.iter
			} else {
				// leave as "false" so it is marked as "new" later
				s.status[result.IP][port.Port] = &openPort{seentAt: s.iter, open: false}
			}
		}
	}
	for ip, ports := range s.status {
		var newPorts []int
		var closedPorts []int
		var totalPorts []int
		var output []string
		report := false

		for port, data := range ports {
			if !data.open && data.seentAt == s.iter {
				data.open = true
				newPorts = append(newPorts, port)
				totalPorts = append(totalPorts, port)
			} else if data.open {
				if s.iter-data.seentAt >= s.ClosedAfter {
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
			log.Printf("%s (%s) - %s", s.reverseIP[ip], ip, strings.Join(output, " | "))
			s.notify(fmt.Sprintf("== %s (%s) ==\n%s", s.reverseIP[ip], ip, strings.Join(output, "\n")))
		}
	}
	return nil
}

// Scan ...
func (s *Scanner) Scan(sleepInterval int) {
	file, err := ioutil.TempFile("", "scanme.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	s.tempFile = file.Name()

	s.scanArgs = append(
		[]string{
			"-oJ", s.tempFile,
			"--rate", s.RateLimit,
			"-p", "1-65535",
		},
		s.Targets...,
	)

	// should probably make iter circular... but even it ticked every second, it's still 136 years to overflow...
	s.iter = 0

	for {
		s.iter++

		s.runScanner()
		s.parseResults()

		if sleepInterval < 1 {
			break
		}
		time.Sleep(time.Duration(sleepInterval) * time.Second)
	}
}
