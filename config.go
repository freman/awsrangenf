package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	gct "github.com/freman/go-commontypes"
	"github.com/naoina/toml"
)

type Config struct {
	Listen  string `json:"-"`
	URL     gct.URL
	Timeout gct.Duration
	Store   string
	IPv6    bool
	Route   struct {
		Table         int
		Gateway       net.IP
		actualGateway net.IP
	}
	Webhook struct {
		Enabled bool
		Key     string
	}
	Polling struct {
		Enabled  bool
		Interval gct.Duration
	}
}

func parseConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open configuration file due to %v", err)
	}
	defer f.Close()

	config := Config{
		Listen:  ":8080",
		URL:     gct.URL{URL: &url.URL{Scheme: "https", Host: "ip-ranges.amazonaws.com", Path: "/ip-ranges.json"}},
		Timeout: gct.Duration{Duration: time.Minute},
	}
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to parse configuration due to %v", err)
	}

	host, port, err := net.SplitHostPort(config.Listen)
	if err != nil {
		return nil, fmt.Errorf("unable to parse listen address due to %v", err)
	}
	if host == "*" {
		host = ""
	}
	if port == "" {
		port = "8000"
	}
	config.Listen = host + ":" + port

	config.Route.actualGateway = config.Route.Gateway
	if config.Route.Gateway.IsUnspecified() {
		config.Route.actualGateway = DefaultRoute()
	}

	return &config, nil
}

func saveConfig(file string, from interface{}) error {
	f, err := os.OpenFile(file+".tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(from); err != nil {
		os.Remove(file + ".tmp")
		return err
	}

	os.Rename(file, file+".bak")
	os.Rename(file+".tmp", file)
	return nil
}

func parseJSON(file string, into interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	return dec.Decode(into)
}

func saveJSON(file string, from interface{}) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(from)
}
