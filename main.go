package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

const banner string = `
    .___                           .__                   __
  __| _/______  ____ ______   ____ |  |__   ____   ____ |  | __
 / __ |\_  __ \/  _ \\____ \_/ ___\|  |  \_/ __ \_/ ___\|  |/ /
/ /_/ | |  | \(  <_> )  |_> >  \___|   Y  \  ___/\  \___|    <
\____ | |__|   \____/|   __/ \___  >___|  /\___  >\___  >__|_ \
     \/              |__|        \/     \/     \/     \/     \/
`

var (
	cidr          = flag.String("cidr", "", "network/subnet")
	stepCount int = 0
)

type CheckItems struct {
	ip      net.IP
	version string
	target  net.IP
	web     string
}

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
	bool, devName, devAddr := IsContainNetwork(*cidr)
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
