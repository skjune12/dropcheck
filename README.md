# dropcheck

dropcheck is a command line tool for checking IP network connectivity.

This script does...
- Send ICMP packet to gateway Address
    - Currently, this script regards gateway address as first host's address of the specified network, that is, when you specify  the network `172.16.100.0/24`, this script regards `172.16.100.1` as a gateway address. 
- Send ICMP packet to the external server
- Querying DNS record (A record for IPv4, AAAA record for IPv6)
- Open Browser for checking Web Browsing.

## Install

```bash
go get github.com/skjune12/dropcheck
```

## Execution

```bash
dropcheck -cidr 192.168.0.0/24
```
