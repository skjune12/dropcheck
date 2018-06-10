package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	ping "github.com/sparrc/go-ping"
)

func CalculateGWAddr(cidr string) net.IP {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatal(err)
	}

	inc(ipNet.IP)
	return ipNet.IP
}

func CheckIPVersion(str string) string {
	ip := net.ParseIP(str)

	if ip != nil {
		if strings.Contains(str, ".") {
			return "IPv4"
		} else if strings.Contains(str, ":") {
			return "IPv6"
		}
	}

	// if not match
	return "None"
}

func DNSLookup(version, addr string) {
	record, err := net.ResolveIPAddr(version, addr)
	if err != nil {
		errMsg := fmt.Sprintf("%s\n", err)
		PrintFAIL(errMsg)
	} else {
		PrintPASS(record.String())
	}
}

func IsContainNetwork(cidr string) (bool, string, net.IP) {
	var devName string
	var devAddr net.IP

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatal("net.ParseCIDR(cidr)", err)
	}

	devices, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		addrs, err := device.Addrs()
		if err != nil {
			log.Fatal(err)
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				log.Fatal(err)
			}

			if ipnet.Contains(ip) {
				devName = device.Name
				devAddr = ip
				return true, devName, devAddr
			}
		}
	}

	// Not Found
	return false, "", nil
}

func Ping(ip net.IP) {
	pinger, err := ping.NewPinger(ip.String())
	if err != nil {
		log.Fatal(err)
	}

	pinger.Count = 10

	pinger.Interval = time.Millisecond * 100
	pinger.Timeout = time.Second * 5
	pinger.Run()

	stats := pinger.Statistics()

	statMsg := fmt.Sprintf("PacketLoss: %.3f%%", stats.PacketLoss)

	if stats.PacketLoss <= 5 {
		PrintPASS(statMsg)
	} else {
		PrintFAIL(statMsg)
	}
}
