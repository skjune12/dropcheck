package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

func Usage() {
	fmt.Printf("Usage: %s TargetIPAddr\n", os.Args[0])
}
