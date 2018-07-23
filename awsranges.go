package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"sort"
	"strings"
	"sync"
)

type FilteredCache struct {
	m        sync.Mutex
	Prefixes []Prefix
	Filter   []string
}

type Prefix struct {
	IPv6    bool
	Prefix  *net.IPNet
	Region  string
	Service string
}

type Prefixes struct {
	Cache           FilteredCache
	PrefixList      []Prefix
	RegionToService map[string][]string
	ServiceToRegion map[string][]string
}

func unmarshalCIDR(dec *json.Decoder) (*net.IPNet, error) {
	var cidr string
	if err := dec.Decode(&cidr); err != nil {
		return nil, err
	}
	_, prefix, err := net.ParseCIDR(cidr)
	return prefix, err
}

func (p *Prefixes) unmarshalPrefix(dec *json.Decoder) (prefix Prefix, err error) {
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return prefix, err
		}

		if s, isa := tok.(string); isa {
			switch s {
			case "ip_prefix":
				prefix.Prefix, err = unmarshalCIDR(dec)
			case "ipv6_prefix":
				prefix.IPv6 = true
				prefix.Prefix, err = unmarshalCIDR(dec)
			case "region":
				err = dec.Decode(&prefix.Region)
			case "service":
				err = dec.Decode(&prefix.Service)
			}
			if err != nil {
				return prefix, err
			}
		}
	}
	dec.Token() // Discard the }

	p.RegionToService[prefix.Region] = append(p.RegionToService[prefix.Region], prefix.Service)
	p.ServiceToRegion[prefix.Service] = append(p.ServiceToRegion[prefix.Service], prefix.Region)
	return
}

func (p *Prefixes) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if delim, isa := tok.(json.Delim); isa && delim == '[' {
			for dec.More() {
				tok, err := dec.Token()
				if err != nil {
					return err
				}
				if delim, isa := tok.(json.Delim); isa && delim == '{' {
					prefix, err := p.unmarshalPrefix(dec)
					if err != nil {
						return err
					}
					p.PrefixList = append(p.PrefixList, prefix)
				}
			}
			dec.Token() // discard the ]
		}
	}
	return nil
}

func (p *Prefixes) Filter(with []string) []Prefix {
	p.Cache.m.Lock()
	defer p.Cache.m.Unlock()
	if p.Cache.Equals(with) {
		return p.Cache.Prefixes
	}

	wanted := map[*net.IPNet]struct{}{}
	p.Cache.Prefixes = []Prefix{}
	for _, v := range with {
		sp := strings.Split(v, ":")
		for _, prefix := range p.PrefixList {
			if _, got := wanted[prefix.Prefix]; got {
				continue
			}

			want := sp[0] == "*" || sp[0] == prefix.Region
			want = want && sp[1] == "*" || sp[1] == prefix.Service

			if want {
				wanted[prefix.Prefix] = struct{}{}
				p.Cache.Prefixes = append(p.Cache.Prefixes, prefix)
			}
		}
	}
	return p.Cache.Prefixes
}

func deduplicateStrings(a []string) []string {
	seen, i := make(map[string]struct{}, len(a)), 0
	for _, v := range a {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		a[i] = v
		i++
	}
	return a[:i]
}

func ParseAWSIPRanges(ipv6 bool, r io.Reader) (*Prefixes, error) {
	prefixes := Prefixes{
		RegionToService: map[string][]string{},
		ServiceToRegion: map[string][]string{},
	}
	dec := json.NewDecoder(r)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}

		if s, isa := tok.(string); isa && (s == "prefixes" || (s == "ipv6_prefixes" && ipv6)) {
			if err := dec.Decode(&prefixes); err != nil {
				return nil, err
			}
		}
	}

	// Slurp the remaining
	if _, isa := r.(io.Closer); !isa {
		io.Copy(ioutil.Discard, r)
	}

	for region := range prefixes.RegionToService {
		prefixes.RegionToService[region] = deduplicateStrings(prefixes.RegionToService[region])
		sort.Strings(prefixes.RegionToService[region])
	}

	for service := range prefixes.ServiceToRegion {
		prefixes.ServiceToRegion[service] = deduplicateStrings(prefixes.ServiceToRegion[service])
		sort.Strings(prefixes.ServiceToRegion[service])
	}

	return &prefixes, nil
}

func (a FilteredCache) Equals(b []string) bool {
	if a.Filter == nil && b == nil {
		return true
	}

	if a.Filter == nil || b == nil {
		return false
	}

	if len(a.Filter) != len(b) {
		return false
	}

	for i := range a.Filter {
		if a.Filter[i] != b[i] {
			return false
		}
	}

	return true
}
