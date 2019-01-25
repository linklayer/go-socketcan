package test

import (
	"bytes"
	"testing"
	"time"

	"github.com/linklayer/go-socketcan/pkg/socketcan"
)

func openIsotpInterface(t *testing.T, rxID uint32, txID uint32) socketcan.Interface {
	dev, err := socketcan.NewIsotpInterface("vcan0", rxID, txID)
	if err != nil {
		t.Errorf("could not create CAN device: %v", err)
	}
	return dev
}

func closeIsotpInterface(t *testing.T, dev socketcan.Interface) {
	err := dev.Close()
	if err != nil {
		t.Errorf("could not close CAN device: %v", err)
	}
}

func TestIsotpOpenClose(t *testing.T) {
	dev := openIsotpInterface(t, 0x100, 0x200)
	closeIsotpInterface(t, dev)
}

func TestIsotpRecvTimeout(t *testing.T) {
	dev := openIsotpInterface(t, 0x100, 0x200)
	err := dev.SetRecvTimeout(50 * time.Millisecond)
	if err != nil {
		t.Errorf("error setting isotp timeout: %v", err)
	}

	_, err = dev.RecvBuf()
	if err == nil {
		t.Error("did not timeout")
	}
}

func TestIsotpTxRxSingleFrame(t *testing.T) {
	txData := []byte{1, 2, 3, 4, 5, 6, 7}

	txDev := openIsotpInterface(t, 0x100, 0x200)
	rxDev := openIsotpInterface(t, 0x200, 0x100)

	go func() {
		time.Sleep(50 * time.Millisecond)
		err := txDev.SendBuf(txData)
		if err != nil {
			t.Errorf("error sending frame: %v", err)
		}
	}()

	rxData, err := rxDev.RecvBuf()
	if err != nil {
		t.Errorf("error receiving frame: %v", err)
	}

	if !bytes.Equal(rxData, txData) {
		t.Error("sent and received data does not match")
	}

	closeIsotpInterface(t, txDev)
	closeIsotpInterface(t, rxDev)
}

func TestIsotpTxRxMultiFrame(t *testing.T) {
	txData := make([]byte, 4095)
	for i := 0; i < 4095; i++ {
		txData[i] = byte(i)
	}

	txDev := openIsotpInterface(t, 0x100, 0x200)
	rxDev := openIsotpInterface(t, 0x200, 0x100)

	go func() {
		time.Sleep(50 * time.Millisecond)
		err := txDev.SendBuf(txData)
		if err != nil {
			t.Errorf("error sending frame: %v", err)
		}
	}()

	rxData, err := rxDev.RecvBuf()
	if err != nil {
		t.Errorf("error receiving frame: %v", err)
	}

	if !bytes.Equal(rxData, txData) {
		t.Error("sent and received data does not match")
	}

	closeIsotpInterface(t, txDev)
	closeIsotpInterface(t, rxDev)
}

func TestIsotpWrongFunctions(t *testing.T) {
	dev := openIsotpInterface(t, 0x100, 0x200)

	err := dev.SendFrame(socketcan.CanFrame{})
	if err == nil {
		t.Errorf("no error trying to SendFrame with isotp interface")
	}

	_, err = dev.RecvFrame()
	if err == nil {
		t.Errorf("no error trying to RecvFrame with raw interface")
	}

	closeRawInterface(t, dev)
}

func TestIsotpTxPadding(t *testing.T) {
	txData := make([]byte, 10)
	for i := 0; i < 10; i++ {
		txData[i] = byte(i)
	}

	txDev := openIsotpInterface(t, 0x100, 0x200)
	rxDev := openIsotpInterface(t, 0x200, 0x100)
	txDev.SetTxPadding(true, 0xAA)
	rxDev.SetTxPadding(true, 0xBB)

	go func() {
		time.Sleep(50 * time.Millisecond)
		err := txDev.SendBuf(txData)
		if err != nil {
			t.Errorf("error sending frame: %v", err)
		}
	}()

	rxData, err := rxDev.RecvBuf()
	if err != nil {
		t.Errorf("error receiving frame: %v", err)
	}

	if !bytes.Equal(rxData, txData) {
		t.Error("sent and received data does not match")
	}

	closeIsotpInterface(t, txDev)
	closeIsotpInterface(t, rxDev)
}
