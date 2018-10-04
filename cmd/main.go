package main

import (
	troll "github.com/gargath/bridge-troll/pkg"
	flag "github.com/spf13/pflag"
)

var watchFiles = flag.StringArrayP("watchfile", "f", nil, "The file to watch. Can be used multiple times")

func main() {
	flag.Parse()
	troll, err := troll.NewBridgeTroll(*watchFiles)
	if err != nil {
		panic(err)
	}
	sync, err := troll.Start()
	if err != nil {
		panic(err)
	}
	sync.Wait()
}
