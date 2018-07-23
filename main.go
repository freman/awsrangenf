package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rice "github.com/GeertJohan/go.rice"
)

func enviromentString(name, def string) string {
	if tmp := os.Getenv(name); tmp != "" {
		return tmp
	}
	return def
}

func mightFindBox(box *rice.Box, err error) *rice.Box {
	return box
}

func main() {
	ring := &ringWriter{}
	logger := log.New(ring, "", log.LstdFlags)

	flgConfig := flag.String("config", enviromentString("CONFIG", "config.toml"), "Path to the configuration file {ENV: CONFIG}")
	flag.Parse()

	cfg, err := parseConfig(*flgConfig)
	if err != nil {
		logger.Println("Unable load configuration file due to", err)
		return
	}

	app := &app{
		config:     cfg,
		configFile: *flgConfig,
		box:        mightFindBox(rice.FindBox("ui/dist")),
		ring:       ring,
		log:        logger,
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{})

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for signal := range sigs {
			switch signal {
			case syscall.SIGINT, syscall.SIGTERM:
				close(done)
				return
			case syscall.SIGHUP:
				newcfg, err := parseConfig(*flgConfig)
				if err != nil {
					logger.Println("Unable load configuration file due to", err)
				} else {
					go app.Reload(newcfg)
				}
			}
		}
	}()

	go app.Run()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
