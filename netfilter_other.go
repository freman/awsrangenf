// +build !linux

package main

import (
	"bytes"
	"net"
	"sort"
)

var pretendRoutes = []*net.IPNet{}

// DefaultRoute returns a fake default route because we're not in linux
func DefaultRoute() net.IP {
	return net.IP{192, 168, 0, 1}
}

func SetRoutes(a *app) error {
	wanted := a.wantedRoutes()
	for i := 0; i < len(pretendRoutes); i++ {
		oldRoute := pretendRoutes[i]

		idx := sort.Search(len(wanted), func(i int) bool {
			return bytes.Compare(wanted[i].IP, oldRoute.IP) >= 0 && bytes.Compare(wanted[i].Mask, oldRoute.Mask) >= 0
		})
		if idx < len(wanted) && bytes.Compare(wanted[idx].IP, oldRoute.IP) == 0 && bytes.Compare(wanted[idx].Mask, oldRoute.Mask) == 0 {
			wanted = append(wanted[:idx], wanted[idx+1:]...)
			continue
		}
		pretendRoutes[i] = pretendRoutes[len(pretendRoutes)-1]
		pretendRoutes = pretendRoutes[:len(pretendRoutes)-1]
		i--
	}

	for _, v := range wanted {
		pretendRoutes = append(pretendRoutes, v)
	}

	return nil
}
