package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	rice "github.com/GeertJohan/go.rice"
	gct "github.com/freman/go-commontypes"
	"github.com/freman/work"
	"github.com/freman/work/bootstrap"
)

type app struct {
	run        *work.Run
	box        *rice.Box
	config     *Config
	configFile string
	httpServer *http.Server
	prefixes   *Prefixes
	bootstrap  *bootstrap.Bootstrap
	selections []string
	customs    []*gct.Network
	timer      *time.Timer
	log        *log.Logger
	ring       *ringWriter
}

const version = `0.0.1`
const userAgent = `awsrangenf/` + version
const httpDate = time.RFC1123

func (a *app) Run() {
	a.bootstrap = &bootstrap.Bootstrap{}
	a.bootstrap.MkdirAll(a.config.Store, 0755).
		IsWritable(a.store("ip-ranges.json")).
		IsWritable(a.store("selections.json")).
		IsWritable(a.store("customs.json")).
		Add(func(next work.Task) work.Task {
			return work.LabelFunc("update prefixes", func(ctx context.Context) error {
				if err := a.update(); err != nil {
					a.log.Println("Unable to update prefixes due to", err)
					return err
				}
				return next.Execute(ctx)
			})
		}).Add(func(next work.Task) work.Task {
		return work.LabelFunc("load selections", func(ctx context.Context) error {
			if err := parseJSON(a.store("selections.json"), &a.selections); err != nil {
				a.log.Println("Unable to load selections due to", err)
				return err
			}
			return next.Execute(ctx)
		})
	}).Add(func(next work.Task) work.Task {
		return work.LabelFunc("update custom ranges", func(ctx context.Context) error {
			if err := parseJSON(a.store("customs.json"), &a.customs); err != nil {
				a.log.Println("Unable to update custom ranges due to", err)
				return err
			}
			return next.Execute(ctx)
		})
	}).Add(func(next work.Task) work.Task {
		return work.LabelFunc("setup routing table", func(ctx context.Context) error {
			if err := SetRoutes(a); err != nil {
				a.log.Println("Unable to setup routing table due to", err)
				return err
			}
			return next.Execute(ctx)
		})
	})

	if a.config.Polling.Enabled {
		a.timer = time.NewTimer(a.config.Polling.Interval.Duration)
	} else {
		a.timer = time.NewTimer(5 * time.Minute)
		a.timer.Stop()
	}

	a.runServer()
	go a.pollingUpdate()
	go a.performUpdate()
}

func (a *app) pollingUpdate() {
	for _ = range a.timer.C {
		a.log.Println("Polling for new ip-ranges.json")
		a.update()
		SetRoutes(a)
	}
}

func (a *app) performUpdate() {
	var err error
	a.run, err = a.bootstrap.Execute(context.TODO())
	if err != nil {
		a.log.Println("Bootstrap failed to run,", err.Error(), ". Will try again")
	}
}

func (a *app) Shutdown(ctx context.Context) error {
	a.log.Println("Bye")
	if a.httpServer != nil {
		return a.httpServer.Shutdown(ctx)
	}
	return nil
}

func (a *app) Reload(cfg *Config) {
	a.log.Println("Reloading configuration")
	serverRestart := a.config.Listen != cfg.Listen || a.config.Webhook.Enabled != cfg.Webhook.Enabled
	pollingEnabledChanged := a.config.Polling.Enabled != cfg.Polling.Enabled
	pollingIntervalChanged := a.config.Polling.Interval.Duration != cfg.Polling.Interval.Duration
	a.config = cfg

	if pollingEnabledChanged || pollingIntervalChanged {
		if a.config.Polling.Enabled {
			a.log.Println("Polling change detected, enabling polling every", a.config.Polling.Interval.Duration)
			a.timer.Reset(a.config.Polling.Interval.Duration)
		} else {
			a.log.Println("Polling change detected, disabling polling")
			a.timer.Stop()
		}
	}

	if serverRestart && a.httpServer != nil {
		a.log.Println("Restarting embedded httpd")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.httpServer.Shutdown(ctx)
		a.runServer()
	}
}

func (a *app) store(file string) string {
	return filepath.Join(a.config.Store, file)
}

func (a *app) update() error {
	httpClient := &http.Client{
		Timeout: a.config.Timeout.Duration,
	}

	req, err := http.NewRequest(http.MethodGet, a.config.URL.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", `application/json`)

	if stat, _ := os.Stat(a.store("ip-ranges.json")); stat != nil {
		req.Header.Set("If-Modified-Since", stat.ModTime().Format(httpDate))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rdr io.Reader
	switch resp.StatusCode {
	case http.StatusOK:
		a.log.Println("Change detected, downloading new ip-ranges.json")
		file, err := os.OpenFile(a.store("ip-ranges.json"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		rdr = io.TeeReader(resp.Body, file)
	case http.StatusNotModified:
		a.log.Println("No change found, reloading ip-ranges.json")
		rdr, err = os.Open(a.store("ip-ranges.json"))
		if err != nil {
			return err
		}
		defer rdr.(io.Closer).Close()
	default:
		a.log.Println("Unexpected http response:", resp.Status)
		return errors.New("unexpected http response")
	}

	prefixes, err := ParseAWSIPRanges(a.config.IPv6, rdr)
	if err != nil {
		os.Remove(a.store("ip-ranges.json"))
		return err
	}
	a.prefixes = prefixes
	return nil
}

func (a *app) wantedRoutes() []*net.IPNet {
	prefixes := a.prefixes.Filter(a.selections)
	customLen := len(a.customs)
	prefixLen := len(prefixes)
	totalLen := customLen + prefixLen

	wantedRoutes := make([]*net.IPNet, totalLen)
	for i, v := range a.customs {
		wantedRoutes[i] = v.IPNet
	}
	for i, v := range prefixes {
		wantedRoutes[customLen+i] = v.Prefix
	}

	seen, i := make(map[string]struct{}, len(wantedRoutes)), 0
	for _, v := range wantedRoutes {
		if _, ok := seen[v.String()]; ok {
			continue
		}
		seen[v.String()] = struct{}{}
		wantedRoutes[i] = v
		i++
	}

	sort.Slice(wantedRoutes[:i], func(i, j int) bool {
		return wantedRoutes[i].String() < wantedRoutes[j].String()
	})

	return wantedRoutes[:i]
}
