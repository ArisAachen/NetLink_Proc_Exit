package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/sys/unix"
	"log"
	"syscall"
	"time"
)

// #include <linux/connector.h>
// #include <linux/cn_proc.h>
import "C"

func main() {
	sock, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_DGRAM, unix.NETLINK_CONNECTOR)
	if err != nil {
		log.Printf(">>>>>> create socket failed, err: %v", err)
		return
	}

	localSockAddr := &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Pid:    uint32(unix.Getgid()),
		Groups: C.CN_IDX_PROC,
	}

	kernelSockAddr := &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Pid:    0,
		Groups: C.CN_IDX_PROC,
	}

	err = unix.Bind(sock, localSockAddr)
	if err != nil {
		log.Printf(">>>>>> bind socket failed, err: %v", err)
		return
	}

	defer unix.Close(sock)

	err = changListenMode(sock, kernelSockAddr, C.PROC_CN_MCAST_LISTEN)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		err = changListenMode(sock, kernelSockAddr, C.PROC_CN_MCAST_IGNORE)
		if err != nil {
			log.Println(err)
			return
		}
	}()

	go func(fd int, sockAddr *unix.SockaddrNetlink) {
		for {
			pid, err := recvExitPid(fd)
			if pid == 0 || err != nil {
				continue
			}
			log.Printf(">>>>>> exit pid is %v", pid)
		}
	}(sock, kernelSockAddr)

	time.Sleep(5 * time.Minute)
}

func changListenMode(fd int, sockAddr unix.Sockaddr, proto uint32) error {
	cnMsg := CnMsg{
		Id: CbId{
			Idx: C.CN_IDX_PROC,
			Val: C.CN_VAL_PROC,
		},
		Ack: 0,
		Seq: 1,
		Len: uint16(binary.Size(proto)),
	}

	nlMsg := unix.NlMsghdr{
		Len:   unix.NLMSG_HDRLEN + uint32(binary.Size(cnMsg)+binary.Size(proto)),
		Type:  uint16(unix.NLMSG_DONE),
		Flags: 0,
		Seq:   1,
		Pid:   uint32(unix.Getpid()),
	}

	// write buf
	buf := bytes.NewBuffer(make([]byte, 0, nlMsg.Len))
	err := binary.Write(buf, binary.LittleEndian, nlMsg)
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, cnMsg)
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, proto)
	if err != nil {
		return err
	}

	err = unix.Sendmsg(fd, buf.Bytes(), nil, sockAddr, 0)
	if err != nil {
		return err
	}
	return nil
}

func recvExitPid(fd int) (uint32, error) {
	buf := make([]byte, 1024)
	nLen, _, _, _, err := unix.Recvmsg(fd, buf, nil, 0)
	if err != nil {
		return 0, err
	}
	if buf == nil {
		return 0, errors.New("not message")
	}
	if nLen < unix.NLMSG_HDRLEN {
		return 0, errors.New("len is not correct")
	}

	nlMsgSlice, err := syscall.ParseNetlinkMessage(buf[:nLen])
	if err != nil {
		log.Fatal(err)
	}

	msg := &CnMsg{}
	header := &ProcEventHeader{}

	bytBuf := bytes.NewBuffer(nlMsgSlice[0].Data)
	err = binary.Read(bytBuf, binary.LittleEndian, msg)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	err = binary.Read(bytBuf, binary.LittleEndian, header)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	if header.What == C.PROC_EVENT_EXIT {
		event := &ExitProcEvent{}
		err = binary.Read(bytBuf, binary.LittleEndian, event)
		if err != nil {
			log.Print(err)
			return 0, nil
		}
		if event.ProcessTgid == event.ProcessPid && event.ProcessTgid == 3270 {
			log.Print("Net")
		}
	}

	return 0, nil
}
