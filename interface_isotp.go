package socketcan

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func NewIsotpInterface(ifName string, rxID uint32, txID uint32) (SocketCanInterface, error) {
	canIf := SocketCanInterface{}
	canIf.ifType = IF_TYPE_ISOTP

	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_DGRAM, CAN_ISOTP)
	if err != nil {
		return canIf, err
	}

	ifIndex, err := getIfIndex(fd, ifName)
	if err != nil {
		return canIf, err
	}

	addr := &unix.SockaddrCAN{Ifindex: ifIndex, RxID: rxID, TxID: txID}
	if err = unix.Bind(fd, addr); err != nil {
		return canIf, err
	}

	canIf.IfName = ifName
	canIf.SocketFd = fd

	return canIf, nil
}

func (i SocketCanInterface) SendBuf(data []byte) error {
	if (i.ifType != IF_TYPE_ISOTP) {
		return fmt.Errorf("interface is not isotp type")
	}

	_, err := unix.Write(i.SocketFd, data)
	return err
}

func (i SocketCanInterface) RecvBuf() ([]byte, error) {
	if (i.ifType != IF_TYPE_ISOTP) {
		return []byte{}, fmt.Errorf("interface is not isotp type")
	}

	data := make([]byte, 4095)
	len, err := unix.Read(i.SocketFd, data)
	if err != nil {
		return data, err
	}

	// only return the valid bytes (0 to length received)
	return data[:len], nil
}
