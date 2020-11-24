//
// Created by root on 2020/11/19.
//

#include "msg_file.h"

#define MAX_MSG_SIZE 256

int
change_listen_mode(int sock, struct sockaddr_nl sock_nl, int mode) {
    // define
    struct msghdr hdr;
    struct iovec iov;
    struct nlmsghdr nl_hdr;

    // memset
    memset(&hdr, 0, sizeof(hdr));
    memset(&iov, 0, sizeof(iov));
    memset(&nl_hdr, 0, sizeof(nl_hdr));

    // set cn_msg
    struct cn_msg* _msg = (struct cn_msg *) NLMSG_DATA(&nl_hdr);
    int *cn_mode = (int *) _msg->data;
    *cn_mode = (mode == 1 ? PROC_CN_MCAST_LISTEN : PROC_CN_MCAST_IGNORE);
    _msg->id.idx = CN_IDX_PROC;
    _msg->id.val = CN_VAL_PROC;
    _msg->ack = 0;
    _msg->seq = 0;
    _msg->len = sizeof(enum proc_cn_mcast_op);

    // set nlmsghdr
    nl_hdr.nlmsg_len = NLMSG_LENGTH(sizeof(struct cn_msg) + sizeof(enum proc_cn_mcast_op));
    nl_hdr.nlmsg_pid = getpid();
    nl_hdr.nlmsg_flags = 0;
    nl_hdr.nlmsg_type = NLMSG_DONE;
    nl_hdr.nlmsg_seq = 0;

    // set iov
    iov.iov_base = (void *) &nl_hdr;
    iov.iov_len = nl_hdr.nlmsg_len;

    // msg
    hdr.msg_name = (void *) &sock_nl;
    hdr.msg_namelen = sizeof(struct sockaddr_nl);
    hdr.msg_iov = &iov;
    hdr.msg_iovlen = 1;

    int ret = sendmsg(sock, &hdr, 0);
    if (ret == -1) {
        printf("send msg error is %s\n", strerror(errno));
        return -1;
    }
    printf("send msg success... ...\n");
    return 1;
}

int
recv_exit_pid(int sock, struct sockaddr_nl sock_nl) {
    // define
    struct msghdr hdr;
    struct iovec iov;
    struct nlmsghdr nl_hdr;

    // memset
    memset(&hdr, 0, sizeof(hdr));
    memset(&iov, 0, sizeof(iov));
    memset(&nl_hdr, 0, sizeof(nl_hdr));

    iov.iov_base = (void *) &nl_hdr;
    iov.iov_len = NLMSG_SPACE(MAX_MSG_SIZE);

    hdr.msg_name = (void *) &sock_nl;
    hdr.msg_namelen = sizeof(struct sockaddr_nl);
    hdr.msg_iov = &iov;
    hdr.msg_iovlen = 1;

    int ret = recvmsg(sock, &hdr, 0);
    if (ret == -1) {
        printf("recv msg err is %s\n", strerror(errno));
        exit(0);
    } else if (ret == 0) {
        printf("rec is empty\n");
    } else {
        printf("rec is success\n");
        struct cn_msg *_msg = (struct cn_msg *) NLMSG_DATA(&nl_hdr);
        struct proc_event *_event = (struct proc_event *) _msg->data;
        switch (_event->what) {
            case PROC_EVENT_EXIT:
                return _event->event_data.exit.process_pid;
            default:
                return 0;
        }
    }
    return 0;
}