package socketcan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	CAN_RAW   = 1
	CAN_ISOTP = 6

	CAN_RAW_RECV_OWN_MSGS = 4
	SOL_CAN_RAW           = 101
)

type SocketCanInterface struct {
	IfName   string
	SocketFd int
}

func getIfIndex(fd int, ifName string) (int, error) {
	ifNameRaw, err := unix.ByteSliceFromString(ifName)
	if err != nil {
		return 0, err
	}
	if len(ifNameRaw) > 16 {
		return 0, errors.New("maximum ifname length is 16 characters")
	}

	ifReq := ifreqIndex{}
	copy(ifReq.Name[:], ifNameRaw)
	err = ioctlIfreq(fd, &ifReq)
	return ifReq.Index, err
}

type ifreqIndex struct {
	Name  [16]byte
	Index int
}

func ioctlIfreq(fd int, ifreq *ifreqIndex) (err error) {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		unix.SIOCGIFINDEX,
		uintptr(unsafe.Pointer(ifreq)),
	)
	if errno != 0 {
		err = fmt.Errorf("ioctl: %v", errno)
	}
	return
}

func NewRawInterface(ifName string) (SocketCanInterface, error) {
	canIf := SocketCanInterface{}

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

func NewIsotpInterface(ifName string, rxID uint32, txID uint32) (SocketCanInterface, error) {
	canIf := SocketCanInterface{}

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

func (i SocketCanInterface) SendFrame(f CanFrame) error {
	// assemble a SocketCAN frame
	frameBytes := make([]byte, 16)
	// bytes 0-3: arbitration ID
	binary.LittleEndian.PutUint32(frameBytes[0:4], uint32(f.ArbId))
	// byte 4: data length code
	frameBytes[4] = f.Dlc
	// data
	copy(frameBytes[8:], f.Data)

	_, err := unix.Write(i.SocketFd, frameBytes)
	return err
}

func (i SocketCanInterface) RecvFrame() (CanFrame, error) {
	f := CanFrame{}

	// read SocketCAN frame from device
	frameBytes := make([]byte, 16)
	_, err := unix.Read(i.SocketFd, frameBytes)
	if err != nil {
		return f, err
	}

	// bytes 0-3: arbitration ID
	f.ArbId = int(binary.LittleEndian.Uint32(frameBytes[0:4]))
	// byte 4: data length code
	f.Dlc = frameBytes[4]
	// data
	f.Data = make([]byte, 8)
	copy(f.Data, frameBytes[8:])

	return f, nil
}

func (i SocketCanInterface) SendBuf(data []byte) error {
	_, err := unix.Write(i.SocketFd, data)
	return err
}

func (i SocketCanInterface) RecvBuf() ([]byte, error) {
	data := make([]byte, 4095)
	len, err := unix.Read(i.SocketFd, data)
	if err != nil {
		return data, err
	}

	// only return the valid bytes (0 to length received)
	return data[:len], nil
}

func (i SocketCanInterface) SetLoopback(enable bool) error {
	value := 0
	if enable {
		value = 1
	}
	err := unix.SetsockoptInt(i.SocketFd, SOL_CAN_RAW, CAN_RAW_RECV_OWN_MSGS, value)
	return err
}

func (i SocketCanInterface) SetRecvTimeout(timeout time.Duration) error {
	tv := unix.Timeval{
		Sec:  int64(timeout / time.Second),
		Usec: int64(timeout % time.Second / time.Microsecond),
	}
	err := unix.SetsockoptTimeval(i.SocketFd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv)
	return err
}

func (i SocketCanInterface) SetSendTimeout(timeout time.Duration) error {
	tv := unix.Timeval{
		Sec:  int64(timeout / time.Second),
		Usec: int64(timeout % time.Second / time.Microsecond),
	}
	err := unix.SetsockoptTimeval(i.SocketFd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, &tv)
	return err
}

func (i SocketCanInterface) Close() error {
	return unix.Close(i.SocketFd)
}
