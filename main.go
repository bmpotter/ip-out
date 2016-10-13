package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"text/tabwriter"
)

// DefaultExcludes declares the default interface exclusions from output
const DefaultExcludes = `^veth|docker|br-`

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

func interfaces(nameExcludePattern *regexp.Regexp) ([]*Interface, error) {
	var out []*Interface

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if nameExcludePattern != nil && nameExcludePattern.MatchString(iface.Name) ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// make sure this is always at least empty so it can be iterated over safely
		addrNames := make([]string, 0)

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			addrNames = append(addrNames, addr.String())
		}

		out = append(out, &Interface{
			Name:      iface.Name,
			LinkUp:    iface.Flags&net.FlagUp != 0,
			Addresses: addrNames,
		})
	}

	return out, nil
}

func main() {
	pp := flag.Bool("pp", false, "Pretty-print output")
	flag.Parse()

	excludeRegexp := regexp.MustCompile(DefaultExcludes)

	outs, err := interfaces(excludeRegexp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading interfaces: %s\n", err)
		os.Exit(1)
	}

	if *pp {
		const padding = 10
		w := tabwriter.NewWriter(os.Stdout, 10, 2, 10, ' ', 0)
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
