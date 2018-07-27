package socketcan

import (
	"fmt"

	"golang.org/x/sys/unix"
)

const CAN_ISOTP_OPTS = 1
const CAN_ISOTP_TX_PADDING = 0x004
const CAN_ISOTP_RX_PADDING = 0x008

func NewIsotpInterface(ifName string, rxID uint32, txID uint32) (Interface, error) {
	canIf := Interface{}
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

func (i Interface) SendBuf(data []byte) error {
	if i.ifType != IF_TYPE_ISOTP {
		return fmt.Errorf("interface is not isotp type")
	}

	_, err := unix.Write(i.SocketFd, data)
	return err
}

func (i Interface) RecvBuf() ([]byte, error) {
	if i.ifType != IF_TYPE_ISOTP {
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

func (i Interface) SetTxPadding(on bool) error {
	// TODO
	return nil
}
