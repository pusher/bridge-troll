package main

import (
	"fmt"
	"os"

	troll "github.com/pusher/bridge-troll/pkg"
	flag "github.com/spf13/pflag"
)

// TODO: Maybe use StringSliceP?
var watchFiles = flag.StringArrayP("watchfile", "f", nil, "The file to watch. Can be used multiple times")
var port = flag.IntP("metrics-port", "p", 2112, "The metrics port to use")
var interval = flag.IntP("check-interval", "i", 10, "The numer of seconds between checks of the watchfile contents")
var metricsPath = flag.StringP("metrics-path", "m", "/metrics", "The path for the metrics endpoint")
var help = flag.BoolP("help", "h", false, "This help")

func main() {
	flag.Parse()
	if *help {
		Usage()
		os.Exit(0)
	}
	troll, err := troll.NewBridgeTroll(*watchFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize BridgeTroll: %v\n\n", err)
		Usage()
		os.Exit(1)
	}
	sync, err := troll.Start(port, metricsPath, interval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start BridgeTroll: %v\n\n", err)
		os.Exit(1)
	}
	sync.Wait()
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	flag.PrintDefaults()
}
