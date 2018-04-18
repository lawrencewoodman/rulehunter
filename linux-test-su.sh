#!/usr/bin/env bash
# This script is used to run the tests under linux as root
#
# Usage:
#    linux-test-su.sh goPath goBinPath initSystem
#
# goPath is the standard GOPATH
# goBinPath is the location of go
# initSystem can be systemd or upstart
#
# Typical usages:
#    sudo ./linux-test-su.sh $GOPATH `which go` systemd
#    sudo ./linux-test-su.sh $GOPATH `which go` upstart
#
# Using Fish:
#    sudo ./linux-test-su.sh $GOPATH (which go) systemd
#    sudo ./linux-test-su.sh $GOPATH (which go) upstart


export GOPATH=$1
export GOBIN=`dirname $(dirname $2)`/bin
$GOBIN/go test -v -tags="su $3" ./...
