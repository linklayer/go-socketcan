package socketcan

import (
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

const (
	IF_TYPE_RAW   = 0
	IF_TYPE_ISOTP = 1
)

type SocketCanInterface struct {
	IfName   string
	SocketFd int
	ifType   int
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
