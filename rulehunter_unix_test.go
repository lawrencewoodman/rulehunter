// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import "fmt"

// Users that are known on a Unix system to try using
var knownUsers = []string{"", "root"}
