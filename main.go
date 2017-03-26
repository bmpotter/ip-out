package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"text/tabwriter"
	"time"

	"syscall"

	// need this to get layer 2 link status from an interface, see https://codereview.appspot.com/105440043/
	"github.com/vishvananda/netlink"
)

const (
	// ipv4 wait timeout seconds
	ipv4WaitTimeoutS = 30

	// defaultExcludes declares the default interface exclusions from output
	defaultExcludes = `^veth|docker|br-|sit`
)

type address []string

func (a address) String() string {
	var outStr string
	if len(a) == 0 {
		outStr = "--"
	} else {
		outStr = ""
		for ix, add := range a {
			if ix > 0 {
				outStr += ", "
			}
			outStr += add
		}
	}

	return fmt.Sprintf("%v", outStr)
}

type link bool

func (l link) String() string {
	if l {
		return fmt.Sprintf("up")
	}

	return fmt.Sprintf("down")
}

// Interface contains human-friendly facts about a network interface
type Interface struct {
	Name      string  `json:"name"`
	LinkUp    link    `json:"link_up"`
	Addresses address `json:"addresses"`
}

func interfaces(nameExcludePattern *regexp.Regexp, ipv4Wait *bool) ([]*Interface, error) {

	gatherAddrs := func(addrs []net.Addr, names *[]string) bool {
		containsIpv4 := false

		for _, addr := range addrs {
			addrString := addr.String()

			if ip, _, err := net.ParseCIDR(addrString); err == nil && ip != nil {
				*names = append(*names, addrString)

				if v4 := ip.To4(); v4 != nil {
					containsIpv4 = true
				}
			}

		}

		return containsIpv4
	}

	var out []*Interface

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || (nameExcludePattern != nil && nameExcludePattern.MatchString(iface.Name)) {
			continue
		}

		// TODO: consider making this a tri-state type so we can reflect unknown / unfetchable value
		var linkUp link
		var link netlink.Link
		if link, err = netlink.LinkByName(iface.Name); err != nil {
			return nil, err
		}

		linkUp = (*link.Attrs()).RawFlags&syscall.IFF_RUNNING == uint32(syscall.IFF_RUNNING)

		addrNames := []string{}

		// only bother gathering addresses if link is up
		if linkUp {
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, err
			}

			start := time.Now()
			for {
				var containsIpv4 bool
				containsIpv4 = gatherAddrs(addrs, &addrNames)

				if ipv4Wait == nil || !*ipv4Wait || containsIpv4 || time.Now().Sub(start) > time.Duration(ipv4WaitTimeoutS)*time.Second {
					break
				}

				time.Sleep(500 * time.Millisecond)
			}
		}

		out = append(out, &Interface{
			Name:      iface.Name,
			LinkUp:    linkUp,
			Addresses: addrNames,
		})
	}

	return out, nil
}

func main() {
	pp := flag.Bool("pp", false, "Pretty-print output")
	ipv4Wait := flag.Bool("ipv4-wait", true, fmt.Sprintf("Wait for at least one ipv4 address to be assigned to an interface before reporting up to timeout of %vs", ipv4WaitTimeoutS))
	flag.Parse()

	excludeRegexp := regexp.MustCompile(defaultExcludes)

	outs, err := interfaces(excludeRegexp, ipv4Wait)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading interfaces: %s\n", err)
		os.Exit(1)
	}

	if *pp {
		const padding = 10
		w := tabwriter.NewWriter(os.Stdout, 8, 2, 8, ' ', 0)
		for _, o := range outs {
			fmt.Fprintf(w, "%v\t%v\t%v\t\n", o.Name, o.LinkUp, o.Addresses)
		}
		w.Flush()

	} else {
		// default is to output JSON
		enc, err := json.Marshal(outs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error serializing output: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf(string(enc))
	}
}
