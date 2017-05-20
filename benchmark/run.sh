#!/bin/sh

sleep 1s

compare () {
    echo ""
    echo ">>>>>>>>>>>>>>>>>>>>"
    echo "$1 > no proxy"
    echo ">>>>>>>>>>>>>>>>>>>>"
    netperf -t $1 -H netserver -p 12865 -v 2 -- -P 12345,8081
    echo "<<<<<<<<<<<<<<<<<<<<"

    sleep 10s

    echo ""
    echo ">>>>>>>>>>>>>>>>>>>>"
    echo "$1 > proxyd"
    echo ">>>>>>>>>>>>>>>>>>>>"
    netperf -t $1 -H proxyd -p 8080 -v 2 -- -P 12345,8081
    echo "<<<<<<<<<<<<<<<<<<<<"

    sleep 10s

    echo ""
    echo ">>>>>>>>>>>>>>>>>>>>"
    echo "$1 > nginx"
    echo ">>>>>>>>>>>>>>>>>>>>"
    netperf -t $1 -H nginx -p 8080 -v 2 -- -P 12345,8081
    echo "<<<<<<<<<<<<<<<<<<<<"
}

compare TCP_RR
sleep 10s
compare TCP_STREAM
sleep 10s
compare TCP_MAERTS
sleep 10s
compare TCP_SENDFILE
sleep 10s
compare TCP_CC
sleep 60s
compare TCP_CRR
