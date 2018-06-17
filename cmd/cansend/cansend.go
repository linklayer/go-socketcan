package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"strconv"

	"github.com/linklayer/go-socketcan/pkg/socketcan"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [interface] [arbid]#[data]\n",
			os.Args[0])
	flag.PrintDefaults()
}

func parseFrame(frameStr string) (socketcan.CanFrame, error) {
	frame := socketcan.CanFrame{}
	fields := strings.Split(frameStr, "#")
	arbId, err := strconv.ParseUint(fields[0], 16, 32)
	if err != nil {
		return frame, err
	}
	frame.ArbId = uint32(arbId)

	if len(fields[1]) % 2 != 0 {
		return frame, fmt.Errorf("invalid frame bytes")
	}
	frame.Dlc = byte(len(fields[1]) / 2)

	frame.Data = make([]byte, frame.Dlc)
	for i := byte(0); i < frame.Dlc; i++ {
		var val, err = strconv.ParseInt(fields[1][i*2:i*2+2], 16, 9)
		if err != nil {
			return frame, err
		}
		frame.Data[i] = byte(val)
	}

	return frame, nil
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) != 2 {
		usage()
		os.Exit(1)
	}

	canIf := flag.Args()[0]
	frameStr := flag.Args()[1]

	frame, err := parseFrame(frameStr)
	if err != nil {
		fmt.Printf("could not parse frame: %v\n", err)
		os.Exit(1)
	}

	var device socketcan.Interface
	device, err = socketcan.NewRawInterface(canIf)
	if err != nil {
		fmt.Printf("could not open interface %s: %v\n",
			canIf, err)
		os.Exit(1)
	}
	defer device.Close()

	device.SendFrame(frame)
}
