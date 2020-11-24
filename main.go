package main

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"log"
	"time"
	"unsafe"
)

// #include "msg_file.h"
// #include <linux/cn_proc.h>
import "C"

const (
	procListen = 1 + iota
	procIgnore
)

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

	// defer sendMsg(sock, kernelSockAddr, 2)

	//err = changeListenMode(sock, kernelSockAddr, procListen)
	//if err != nil {
	//
	//}
	//
	//// defer changeListenMode(sock, kernelSockAddr, procIgnore)
	//
	go func(fd int, sockAddr *unix.SockaddrNetlink) {
		for {
			pid := recvExitPid(fd, sockAddr)
			log.Printf(">>>>>> exit pid is %v", pid)
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
	err := binary.Write(buf, binary.BigEndian, nlMsg)
	if err != nil {
		log.Printf(">>>>>> write nl proto failed, err: %v \n", err)
		return err
	}

	err = binary.Write(buf, binary.BigEndian, cnMsg)
	if err != nil {
		log.Printf(">>>>>> write nl cn msg failed, err: %v \n", err)
		return err
	}

	err = binary.Write(buf, binary.BigEndian, uint32(proto))
	if err != nil {
		log.Printf(">>>>>> write nl cn msg failed, err: %v \n", err)
		return err
	}

	err = unix.Sendmsg(fd, buf.Bytes(), nil, sockAddr, 0)
	if err != nil {
		log.Printf(">>>>>> send msg failed, err: %v \n", err)
		return err
	}
	return nil
}

//func changeListenMode(fd int, sockAddr *unix.SockaddrNetlink, mode int) error {
//	addr := *(*C.struct_sockaddr_nl)(unsafe.Pointer(sockAddr))
//	ret := C.change_listen_mode(C.int(fd), addr, C.int(mode))
//	if ret == -1 {
//		return errors.New("change mode failed")
//	}
//	return nil
//}

func recvExitPid(fd int, sock *unix.SockaddrNetlink) int {
	pid := C.recv_exit_pid(C.int(fd), *(*C.struct_sockaddr_nl)(unsafe.Pointer(sock)))
	return int(pid)
}
