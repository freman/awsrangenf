package main

import (
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
)

var sampleJson = `{
	"syncToken": "1531345951",
	"createDate": "2018-07-11-21-52-31",
	"prefixes": [
		{
			"ip_prefix": "18.208.0.0/13",
			"region": "us-east-1",
			"service": "AMAZON"
		},
		{
			"ip_prefix": "52.95.245.0/24",
			"region": "us-east-1",
			"service": "AMAZON"
		}
	],
	"ipv6_prefixes": [
		{
			"ipv6_prefix": "2600:1f18::/33",
			"region": "us-east-1",
			"service": "EC2"
		},
		{
			"ipv6_prefix": "2600:1fff:5000::/40",
			"region": "us-gov-east-1",
			"service": "EC2"
		}
	]
}`

func TestParseAWSIPRanges(t *testing.T) {
	expect := &Prefixes{
		PrefixList: []Prefix{
			Prefix{IPv6: false, Prefix: &net.IPNet{IP: net.IP{0x12, 0xd0, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xf8, 0x0, 0x0}}, Region: "us-east-1", Service: "AMAZON"},
			Prefix{IPv6: false, Prefix: &net.IPNet{IP: net.IP{0x34, 0x5f, 0xf5, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}}, Region: "us-east-1", Service: "AMAZON"},
		},
		RegionToService: map[string][]string{"us-east-1": {"AMAZON"}},
		ServiceToRegion: map[string][]string{"AMAZON": {"us-east-1"}},
	}

	r, err := ParseAWSIPRanges(false, strings.NewReader(sampleJson))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expect, r) {
		t.Errorf("Expected %+v to equal %+v", r, expect)
	}
}

func TestParseIPv6AWSIPRanges(t *testing.T) {
	expect := &Prefixes{
		PrefixList: []Prefix{
			Prefix{IPv6: false, Prefix: &net.IPNet{IP: net.IP{0x12, 0xd0, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xf8, 0x0, 0x0}}, Region: "us-east-1", Service: "AMAZON"},
			Prefix{IPv6: false, Prefix: &net.IPNet{IP: net.IP{0x34, 0x5f, 0xf5, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}}, Region: "us-east-1", Service: "AMAZON"},
			Prefix{IPv6: true, Prefix: &net.IPNet{IP: net.IP{0x26, 0x0, 0x1f, 0x18, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}}, Region: "us-east-1", Service: "EC2"},
			Prefix{IPv6: true, Prefix: &net.IPNet{IP: net.IP{0x26, 0x0, 0x1f, 0xff, 0x50, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}}, Region: "us-gov-east-1", Service: "EC2"},
		},
		RegionToService: map[string][]string{"us-east-1": {"AMAZON", "EC2"}, "us-gov-east-1": {"EC2"}},
		ServiceToRegion: map[string][]string{"EC2": {"us-east-1", "us-gov-east-1"}, "AMAZON": {"us-east-1"}},
	}

	r, err := ParseAWSIPRanges(true, strings.NewReader(sampleJson))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expect, r) {
		t.Errorf("Expected %+v to equal %+v", r, expect)
	}
}

func TestParseRealRanges(t *testing.T) {
	f, err := os.Open("ip-ranges.json")
	if err != nil {
		t.Skip("Unable to open ip-ranges.json")
	}
	defer f.Close()

	prefixes, err := ParseAWSIPRanges(true, f)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(prefixes.PrefixList) == 0 {
		t.Error("No prefixes loaded")
	}
}
