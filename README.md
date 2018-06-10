# dropcheck

dropcheck is a command line tool for diagnosing IPv4/IPv6 network connectivity.

This script does...
- Send ICMP packet to the gateway Address
    - Currently, this script regards a gateway address as first host address of the specified network: that is, when you specify the subnet `172.16.100.0/24`, this script regards `172.16.100.1` as a gateway address. It is also same as IPv6. (When you specify `fc00:dead:beef:1234::/64`, the gateway adress will be `fc00:dead:beef:1234::1`)
- Send ICMP packet to the external server
- Querying DNS record (A record for IPv4, AAAA record for IPv6)
- Open Browser for diagnosing web browsing.

## Install

```bash
go get github.com/skjune12/dropcheck
```

After running `go get`, the execution file will be created in `$GOPATH/bin` if you set $GOPATH collectly.

## Execution

### For IPv4

```bash
dropcheck -cidr 172.16.100.0/24
```

### For IPv6

```bash
dropcheck -cidr fc00:dead:beef:1234::/64
```

## Contact

Kohei Suzuki <jingle@sfc.wide.ad.jp>
