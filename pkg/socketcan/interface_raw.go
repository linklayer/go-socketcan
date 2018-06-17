package socketcan

import (
	"fmt"
	"encoding/binary"

	"golang.org/x/sys/unix"
)

func NewRawInterface(ifName string) (Interface, error) {
	canIf := Interface{}
	canIf.ifType = IF_TYPE_RAW

	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_RAW, CAN_RAW)
	if err != nil {
		return canIf, err
	}

	ifIndex, err := getIfIndex(fd, ifName)
	if err != nil {
		return canIf, err
	}

	addr := &unix.SockaddrCAN{Ifindex: ifIndex}
	if err = unix.Bind(fd, addr); err != nil {
		return canIf, err
	}

	canIf.IfName = ifName
	canIf.SocketFd = fd

	return canIf, nil
}

func (i Interface) SendFrame(f CanFrame) error {
	if (i.ifType != IF_TYPE_RAW) {
		return fmt.Errorf("interface is not raw type")
	}

	// assemble a SocketCAN frame
	frameBytes := make([]byte, 16)
	// bytes 0-3: arbitration ID
	binary.LittleEndian.PutUint32(frameBytes[0:4], f.ArbId)
	// byte 4: data length code
	frameBytes[4] = f.Dlc
	// data
	copy(frameBytes[8:], f.Data)

	_, err := unix.Write(i.SocketFd, frameBytes)
	return err
}

func (i Interface) RecvFrame() (CanFrame, error) {
	f := CanFrame{}

	if (i.ifType != IF_TYPE_RAW) {
		return f, fmt.Errorf("interface is not raw type")
	}

	// read SocketCAN frame from device
	frameBytes := make([]byte, 16)
	_, err := unix.Read(i.SocketFd, frameBytes)
	if err != nil {
		return f, err
	}

	// bytes 0-3: arbitration ID
	f.ArbId = uint32(binary.LittleEndian.Uint32(frameBytes[0:4]))
	// byte 4: data length code
	f.Dlc = frameBytes[4]
	// data
	f.Data = make([]byte, 8)
	copy(f.Data, frameBytes[8:])

	return f, nil
}
