package main

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"log"
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

	sendMsg(sock, kernelSockAddr, C.PROC_CN_MCAST_LISTEN)

	go func(fd int, sockAddr *unix.SockaddrNetlink) {
		for {
			var buf []byte
			_, _, _, _, err := unix.Recvmsg(fd, buf, nil, 0)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(">>>>>> print success")
		}
	}(sock, kernelSockAddr)

	time.Sleep(5 * time.Minute)
}

func sendMsg(fd int, sockAddr unix.Sockaddr, proto int) error {

	cnMsg := CnMsg{
		Id: CbId{
			Idx: C.CN_IDX_PROC,
			Val: C.CN_VAL_PROC,
		},
		Ack: 0,
		Seq: 0,
		Len: uint16(binary.Size(uint32(proto))),
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
		log.Printf(">>>>>> write nl proto failed, err: %v \n", err)
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, cnMsg)
	if err != nil {
		log.Printf(">>>>>> write nl cn msg failed, err: %v \n", err)
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, uint32(proto))
	if err != nil {
		log.Printf(">>>>>> write nl cn msg failed, err: %v \n", err)
		return err
	}

	err = unix.Sendto(fd, buf.Bytes(), 0, sockAddr)
	if err != nil {
		log.Printf(">>>>>> send msg failed, err: %v \n", err)
		return err
	}
	return nil
}
