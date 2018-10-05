package main

import (
	"fmt"
	"os"

	troll "github.com/gargath/bridge-troll/pkg"
	flag "github.com/spf13/pflag"
)

var watchFiles = flag.StringArrayP("watchfile", "f", nil, "The file to watch. Can be used multiple times")
var port = flag.IntP("metrics-port", "p", 2112, "The metrics port to use")
var help = flag.BoolP("help", "h", false, "This help")

func main() {
	flag.Parse()
	if *help {
		Usage()
		os.Exit(0)
	}
	troll, err := troll.NewBridgeTroll(*watchFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start BridgeTroll: %v\n\n", err)
		Usage()
		os.Exit(1)
	}
	sync, err := troll.Start(port)
	if err != nil {
		panic(err)
	}
	sync.Wait()
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	flag.PrintDefaults()
}
