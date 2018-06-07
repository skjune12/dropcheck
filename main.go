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

func PrintStep() {
	stepCount++
	bold := color.New(color.Bold)
	bold.Printf("[Step%d]\t", stepCount)
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
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatal(err)
	}

	inc(ipNet.IP)
	return ipNet.IP
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
	vlanId        = flag.Int("vlan", 0, "vlan-id (Currently no need to specify any number.)")
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

	// NOTE: temporary comment out since we don't use vlanId currently.
	/*
		if *vlanId == 0 {
			fmt.Printf("Enter vlan-id: ")

			input, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			input = strings.TrimSuffix(input, "\n")

			tmp, err := strconv.Atoi(input)
			if err != nil {
				log.Fatal(err)
			}
			vlanId = &tmp
		}
	*/
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
		fmt.Println("Invalid CIDR Format: xxx.xxx.xxx.xxx/netmask")
		os.Exit(1)
	}

	fmt.Println(banner)

	// Check
	PrintStep()
	fmt.Printf("Check if you are connected to the specified network.\n")
	bool, devName, devAddr := IsContainNet(*cidr)
	if bool != true {
		PrintFAIL()
		fmt.Printf("You are not connected to the specified network!\n")
		os.Exit(1)
	} else {
		PrintPASS()
		fmt.Printf("devName: %s, addr: %v\n", devName, devAddr)
	}

	// ping to the gateway addr
	gwAddr := CalculateGWAddr(*cidr)
	PrintStep()
	fmt.Printf("Send ICMP Packet to the gateway addr (destination: %s)\n", gwAddr.String())
	Ping(gwAddr)

	// If IPv6, ping to the link local address
	// TODO: Specify Source Interface

	// if IsIPv6(item.ip.String()) {
	// 	linkLocalGWAddr := CalculateLinkLocalAddr(*vlanId)
	// 	PrintStep(1)
	// 	fmt.Printf("Send ICMP Packet to the link-local gateway addr (destination: %s)\n", linkLocalGWAddr.String())
	// 	Ping(linkLocalGWAddr)
	// }

	// ping to the Internet
	PrintStep()
	fmt.Printf("Send ICMP Packet to the Internet (destination: %s)\n", item.target.String())
	Ping(item.target)

	// Query DNS
	PrintStep()
	fmt.Printf("Query DNS record of 'www.wide.ad.jp'\n")
	DNSLookup(item.version, "www.wide.ad.jp")
	time.Sleep(1000 * time.Millisecond)

	// Open Website
	PrintStep()
	fmt.Printf("Web Browsing (%s)\n", item.web)
	PrintWARN()
	fmt.Printf("Please check your browser.\n")
	open.Run(item.web)

	// Show Result
	fmt.Println("")
	fmt.Println("Finish Checking All The Steps.")
}
