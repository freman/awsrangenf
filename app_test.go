package main

import (
	"net"
	"reflect"
	"testing"

	gct "github.com/freman/go-commontypes"
)

func TestWantedRoutes(t *testing.T) {
	a := &app{
		prefixes: &Prefixes{
			PrefixList: []Prefix{
				Prefix{IPv6: false, Prefix: &net.IPNet{IP: net.IP{0x12, 0xd0, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xf8, 0x0, 0x0}}, Region: "us-east-1", Service: "AMAZON"},
				Prefix{IPv6: false, Prefix: &net.IPNet{IP: net.IP{0x34, 0x5f, 0xf5, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}}, Region: "us-east-1", Service: "AMAZON"},
			},
			RegionToService: map[string][]string{"us-east-1": {"AMAZON"}},
			ServiceToRegion: map[string][]string{"AMAZON": {"us-east-1"}},
		},
		selections: []string{"us-east-1:*"},
		customs: []*gct.Network{
			&gct.Network{IPNet: &net.IPNet{IP: net.IP{0xa, 0xa, 0xa, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}}},
			&gct.Network{IPNet: &net.IPNet{IP: net.IP{0xc0, 0xa8, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}}},
		},
	}

	got := a.wantedRoutes()
	expected := []*net.IPNet{
		&net.IPNet{IP: net.IP{0xa, 0xa, 0xa, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}},
		&net.IPNet{IP: net.IP{0x12, 0xd0, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xf8, 0x0, 0x0}},
		&net.IPNet{IP: net.IP{0xc0, 0xa8, 0x0, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}},
		&net.IPNet{IP: net.IP{0x34, 0x5f, 0xf5, 0x0}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v got %v", expected, got)
	}
}
