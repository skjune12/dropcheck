package main

import (
	"bufio"
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

func PrintPASS(text string) {
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)
	boldGreen.Printf("> PASS\t")
	fmt.Printf("%s\n", text)
}

func PrintFAIL(text string) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("> FAIL\t")
	fmt.Printf("%s\n", text)

}

func PrintWARN(text string) {
	red := color.New(color.FgYellow)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("> WARN\t")
	fmt.Printf("%s\n", text)
}

func PrintStep(text string) {
	stepCount++
	bold := color.New(color.Bold)
	bold.Printf("[Step%d]\t", stepCount)
	fmt.Printf("%s\n", text)
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

	statMsg := fmt.Sprintf("PacketLoss: %.3f%%", stats.PacketLoss)

	if stats.PacketLoss <= 5 {
		PrintPASS(statMsg)
	} else {
		PrintFAIL(statMsg)
	}
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
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatal(err)
	}

	inc(ipNet.IP)
	return ipNet.IP
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
		errMsg := fmt.Sprintf("%s\n", err)
		PrintFAIL(errMsg)
	} else {
		PrintPASS(record.String())
	}
}

func IsContainNet(cidr string) (bool, string, net.IP) {
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

func Usage() {
	fmt.Printf("Usage: %s TargetIPAddr\n", os.Args[0])
}

var (
	cidr          = flag.String("cidr", "", "network/subnet")
	stepCount int = 0
)

func init() {
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	if *cidr == "" {
		fmt.Printf("Enter CIDR: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSuffix(input, "\n")
		cidr = &input
	}
}

type CheckItems struct {
	ip      net.IP
	version string
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
			net.ParseIP("2001:4860:4860::8888"),
			"https://ipv6.google.com"}

	default:
		fmt.Println("Invalid CIDR Format: xxx.xxx.xxx.xxx/netmask")
		os.Exit(1)
	}

	fmt.Println(banner)

	// Check
	PrintStep("Check if you are connected to the specified network.")
	bool, devName, devAddr := IsContainNet(*cidr)
	if bool != true {
		PrintFAIL("You are not connected to the specified network!")
		os.Exit(1)
	} else {
		passMsg := fmt.Sprintf("devName: %s, addr: %v", devName, devAddr)
		PrintPASS(passMsg)
	}

	// ping to the gateway addr
	gwAddr := CalculateGWAddr(*cidr)
	text := fmt.Sprintf("Send ICMP Packet to the gateway addr (destination: %s)", gwAddr.String())
	PrintStep(text)
	Ping(gwAddr)

	// ping to the Internet
	text = fmt.Sprintf("Send ICMP Packet to the Internet (destination: %s)", item.target.String())
	PrintStep(text)
	Ping(item.target)

	// Query DNS
	PrintStep("Query DNS record of 'www.wide.ad.jp'")
	DNSLookup(item.version, "www.wide.ad.jp")

	// Open Website
	text = fmt.Sprintf("Web Browsing (%s)", item.web)
	PrintStep(text)
	PrintWARN("Please check your browser.")
	open.Run(item.web)

	// Show Result
	fmt.Println("")
	fmt.Println("Finish Checking All The Steps.")
}
