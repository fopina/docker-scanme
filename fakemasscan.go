package main

import (
	"fmt"
	"log"
	"flag"
	"os"
    "strings"
    "io/ioutil"
)

func main() {
	fileNamePtr := flag.String("oJ", "/dev/null", "output filename")
    setupPtr := flag.Bool("setup", false, "setup fake port results")
    _ = flag.String("rate", "100", "fake param")
    _ = flag.String("p", "80", "fake param")

	flag.Parse()

    if *setupPtr {
        fileName := os.Getenv("FAKEMASSCAN")
        file, err := os.Create(fileName)
        if err != nil {
            log.Fatal("Cannot create file", err)
        }
        defer file.Close()
        fmt.Fprintf(file, strings.Join(flag.Args(), " "))
    } else {
        file, err := os.Create(*fileNamePtr)
        if err != nil {
            log.Fatal("Cannot create file", err)
        }
        defer file.Close()

        log.Printf("Parameters passed: %s\n", flag.Args())

        fileName := os.Getenv("FAKEMASSCAN")
        ports := ""

        if fileName != "" {
            b, err := ioutil.ReadFile(fileName)
            if err != nil {
                log.Fatal(err)
            }
            str := string(b)
            runsArray := strings.Split(str, " ")
            ports = runsArray[0]

            ioutil.WriteFile(fileName, []byte(strings.Join(runsArray[1:], " ")), 0666)
        }

        if ports == "" {
            fmt.Fprintf(file, "")
        } else {
            portArray := strings.Split(ports, ",")
            output := make([]string, len(portArray))

            for pi, port := range portArray {
                output[pi] = `{   "ip": "45.33.32.156",   "timestamp": "1546596407", "ports": [ {"port": ` + string(port) + `, "proto": "tcp", "status": "open", "reason": "syn-ack", "ttl": 49} ] }`
            }

            fmt.Fprintf(file, "[" + strings.Join(output, ",") + "]")
        }
    }
}
