package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/linklayer/go-socketcan/pkg/socketcan"
)

func dataToString(data []byte) string {
	str := ""
	for i := 0; i < len(data); i++ {
		str = fmt.Sprintf("%s%02X ", str, data[i])
	}
	return str
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [interface]\n",
			os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) != 1 {
		usage()
		os.Exit(1)
	}

	canIf := flag.Args()[0]

	device, err := socketcan.NewRawInterface(canIf)
	if err != nil {
		fmt.Printf("could not open interface %s: %v\n",
			canIf, err)
		os.Exit(1)
	}
	defer device.Close()

	for {
		frame, err := device.RecvFrame()
		if err != nil {
			fmt.Printf("error receiving frame: %v", err)
			os.Exit(1)
		}
		dataStr := dataToString(frame.Data)
		fmt.Printf(" %s\t%03X\t[%d]\t%s\n", device.IfName, frame.ArbId, frame.Dlc, dataStr)
	}
}
