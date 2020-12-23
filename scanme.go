package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fopina/scanme/scanner"
)

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

	s := scanner.Scanner{
		MasscanPath: *masscanPathPtr, RateLimit: *rateLimitPtr, ShowOutput: *showOutputPtr,
		NotifyToken: *notifyTokenPtr, ClosedAfter: uint(*closedAfter), Targets: flag.Args(),
	}
	s.Scan(*sleepIntervalPtr)
}
