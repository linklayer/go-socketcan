package test

import (
	"reflect"
	"testing"
	"time"

	"github.com/linklayer/go-socketcan"
)

func openRawInterface(t *testing.T) socketcan.SocketCanInterface {
	dev, err := socketcan.NewRawInterface("vcan0")
	if err != nil {
		t.Errorf("could not create CAN device: %v", err)
	}
	return dev
}

func closeRawInterface(t *testing.T, dev socketcan.SocketCanInterface) {
	err := dev.Close()
	if err != nil {
		t.Errorf("could not close CAN device: %v", err)
	}
}

func TestRawOpenClose(t *testing.T) {
	dev := openRawInterface(t)
	closeRawInterface(t, dev)
}

func TestRawLoopback(t *testing.T) {
	tx, rx := socketcan.CanFrame{}, socketcan.CanFrame{}
	tx.ArbId = 0x100
	tx.Dlc = 8
	tx.Data = []byte{1, 2, 3, 4, 5, 6, 7, 8}

	dev := openRawInterface(t)

	err := dev.SetLoopback(true)
	if err != nil {
		t.Errorf("could not enable loopback: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		err = dev.SendFrame(tx)
		if err != nil {
			t.Errorf("error sending frame: %v", err)
		}
	}()

	rx, err = dev.RecvFrame()
	if err != nil {
		t.Errorf("error receiving frame: %v", err)
	}

	if !reflect.DeepEqual(rx, tx) {
		t.Error("sent and received frames do not match")
	}

	closeRawInterface(t, dev)
}

func TestRawRecvTimeout(t *testing.T) {
	dev := openRawInterface(t)

	err := dev.SetRecvTimeout(50 * time.Millisecond)
	if err != nil {
		t.Errorf("error setting raw timeout: %v", err)
	}

	_, err = dev.RecvFrame()
	if err == nil {
		t.Error("did not timeout")
	}

	closeRawInterface(t, dev)
}
