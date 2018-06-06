package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	ping "github.com/sparrc/go-ping"
)

const banner string = `
    .___                           .__                   __
  __| _/______  ____ ______   ____ |  |__   ____   ____ |  | __
 / __ |\_  __ \/  _ \\____ \_/ ___\|  |  \_/ __ \_/ ___\|  |/ /
/ /_/ | |  | \(  <_> )  |_> >  \___|   Y  \  ___/\  \___|    <
\____ | |__|   \____/|   __/ \___  >___|  /\___  >\___  >__|_ \
     \/              |__|        \/     \/     \/     \/     \/
`

func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

func GetLinkLocalAddr(name string) string {
	var linkLocalAddr string
	dev, err := net.InterfaceByName(name)
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := dev.Addrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range addrs {
		if strings.Contains(addr.String(), "fe80:") {
			linkLocalAddr = addr.String()
		}
	}

	return linkLocalAddr
}

func PrintPASS() {
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)
	boldGreen.Printf("> PASS\t")
}

func PrintFAIL() {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("> FAIL\t")
}

func PrintWARN() {
	red := color.New(color.FgYellow)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("> WARN\t")
}

func PrintStep(i int) {
	bold := color.New(color.Bold)
	bold.Printf("[Step%d]\t", i)
}

func Ping(ip net.IP) {
	pinger, err := ping.NewPinger(ip.String())
	if err != nil {
		log.Fatal(err)
	}

	pinger.Count = 10

	// TODO: set host ip addr
	// pinger.source = GetLinkLocalAddr(*devname)

	pinger.Interval = time.Millisecond * 100
	pinger.Timeout = time.Second * 5
	pinger.Run()

	stats := pinger.Statistics()

	if stats.PacketLoss <= 5 {
		PrintPASS()
	} else {
		PrintFAIL()
	}
	fmt.Printf("PacketLoss: %.3f%%\n", stats.PacketLoss)
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

func CalculateGWAddr(cidr string) net.IP {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatal(err)
	}

	inc(ip)
	return ip
}

func CalculateLinkLocalAddr(vlanId int) net.IP {
	addr := fmt.Sprintf("fe80::%d:1", vlanId)
	return net.ParseIP(addr)
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func DNSLookup(version, addr string) {
	record, err := net.ResolveIPAddr(version, addr)
	if err != nil {
		PrintFAIL()
		fmt.Printf("%s\n", err)
	} else {
		PrintPASS()
		fmt.Println(record.String())
	}
}

func Usage() {
	fmt.Printf("Usage: %s TargetIPAddr\n", os.Args[0])
}

var (
	cidr    = flag.String("cidr", "192.168.0.0/16", "network/subnet")
	vlanId  = flag.Int("vlan", 1, "vlan-id")
	devname = flag.String("interface", "en7", "device name")
)

func init() {
	flag.Parse()
}

type CheckItems struct {
	ip      net.IP
	version string
	vlanId  int
	target  net.IP
	web     string
}

func main() {
	var item CheckItems
	addr := strings.Split(*cidr, "/")

	switch CheckIPVersion(addr[0]) {
	case "IPv4":
		ip, _, err := net.ParseCIDR(*cidr)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		item = CheckItems{
			ip,
			"ip",
			*vlanId,
			net.ParseIP("8.8.8.8"),
			"https://ipv4.google.com"}

	case "IPv6":
		ip, _, err := net.ParseCIDR(*cidr)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		item = CheckItems{
			ip,
			"ip6",
			*vlanId,
			net.ParseIP("2001:4860:4860::8888"),
			"https://ipv6.google.com"}

	default:
		Usage()
		os.Exit(1)
	}

	fmt.Println(banner)

	// 1. ping to the gateway addr
	gwAddr := CalculateGWAddr(*cidr)
	PrintStep(1)
	fmt.Printf("Send ICMP Packet to the gateway addr (destination: %s)\n", gwAddr.String())
	Ping(gwAddr)

	// If IPv6, ping to the link local address
	// TODO: Specify Source Interface

	if IsIPv6(item.ip.String()) {
		linkLocalGWAddr := CalculateLinkLocalAddr(*vlanId)
		PrintStep(1)
		fmt.Printf("Send ICMP Packet to the link-local gateway addr (destination: %s)\n", linkLocalGWAddr.String())
		Ping(linkLocalGWAddr)
	}

	// 2. ping to the Internet
	PrintStep(2)
	fmt.Printf("Send ICMP Packet to the Internet (destination: %s)\n", item.target.String())
	Ping(item.target)

	// 3. Query DNS
	PrintStep(3)
	fmt.Printf("Query DNS record of 'www.wide.ad.jp'\n")
	DNSLookup(item.version, "www.wide.ad.jp")

	// 4. Open Website
	PrintStep(4)
	fmt.Printf("Web Browsing (%s)\n", item.web)
	PrintWARN()
	fmt.Printf("Please check your browser.\n")
	open.Run(item.web)

	// Show Result
	fmt.Println("")
	fmt.Println("Finish Checking All The Steps.")
}
