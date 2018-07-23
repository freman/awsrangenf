// +build linux

package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"sync"

	"github.com/vishvananda/netlink"
)

var nfLock sync.Mutex

func DefaultRoute() net.IP {
	list, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		panic(err)
	}

	for _, route := range list {
		if route.Dst == nil && route.Src == nil {
			return route.Gw.To16()
		}
	}

	return net.IP{}
}

func SetRoutes(a *app) error {
	a.log.Println("Refreshing netfilter routes")
	nfLock.Lock()
	defer nfLock.Unlock()

	wanted := a.wantedRoutes()

	existing, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: a.config.Route.Table}, netlink.RT_FILTER_TABLE)
	if err != nil {
		a.log.Println("Failed to retrieve route list from netlink:", err)
		return err
	}

	for _, oldRoute := range existing {
		idx := sort.Search(len(wanted), func(i int) bool {
			return wanted[i].String() >= oldRoute.Dst.String()
		})
		if idx < len(wanted) && wanted[idx].String() == oldRoute.Dst.String() {
			wanted = append(wanted[:idx], wanted[idx+1:]...)
			continue
		}
		netlink.RouteDel(&oldRoute)
	}

	for _, v := range wanted {
		err := netlink.RouteAdd(&netlink.Route{
			Table: a.config.Route.Table,
			Dst:   v,
			Gw:    a.config.Route.actualGateway,
		})
		if err != nil && !os.IsExist(err) {
			a.log.Printf("Failed write route %v to netlink: %v", v, err)
			return fmt.Errorf("%v: %v", v, err)
		}
	}
	return nil
}
