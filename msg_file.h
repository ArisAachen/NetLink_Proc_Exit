//
// Created by aris on 2020/11/20.
//

#ifndef C_TEST_MSG_FILE_H
#define C_TEST_MSG_FILE_H

#include <linux/connector.h>
#include <linux/netlink.h>
#include <string.h>
#include <sys/socket.h>
#include <linux/cn_proc.h>
#include <unistd.h>
#include <stdio.h>
#include <errno.h>
#include <stdlib.h>

int
change_listen_mode(int sock, struct sockaddr_nl sock_nl, int mode);

int
recv_exit_pid(int sock, struct sockaddr_nl sock_nl);

#endif //C_TEST_MSG_FILE_H
