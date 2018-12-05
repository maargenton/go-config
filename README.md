# go-config

Configuration file loading and reloading for orhecstrated Go applications.

[![GoDoc](https://godoc.org/github.com/marcus999/go-config?status.svg)](https://godoc.org/github.com/marcus999/go-config)
[![Version](https://img.shields.io/github/tag/marcus999/go-config.svg)](https://github.com/marcus999/go-config)
[![Build Status](https://travis-ci.org/marcus999/go-config.svg?branch=master)](https://travis-ci.org/marcus999/go-config)
[![codecov](https://codecov.io/gh/marcus999/go-config/branch/master/graph/badge.svg)](https://codecov.io/gh/marcus999/go-config)
[![Go Report Card](https://goreportcard.com/badge/github.com/marcus999/go-config)](https://goreportcard.com/report/github.com/marcus999/go-config)


Package `go-config` provide the necessary machinery to load, watch and reload
configuration files for a Go application.

Unlike other packages with similar purpose, `go-config` does not attempt to mix
configuration files with command line and environment variables handling.
In the context of a Kubernetes cluster, confirugation files are handled by
configMap  object that are updated asynchronously from the application
deployment, leading to the following guidelines:

- Command-line arguments and environment variables are attached to the
application deployment definition and cannot be changed without restarting the
application process.
They should be used for configuration requiring the application to restart when
changed. For stateless apllications that are cheap to restart, that should be
almost all the configuration.

- Configuration files may be updated asynchronously from the application
deployment and therefore **MUST** be watched and reloaded when they change.

`go-config` provides the necessary machinery to load and watch configuration
files, and trigger updates notification. It supports YAML and JSON configuration
file format, and loads their content into an application defined config struct.
Reloading is asynchrnous and atomic.

## Installation

    go get github.com/marcus999/go-config

## Usage

```go
package main

import (
	"log"
	"time"

	config "github.com/marcus999/go-config"
)

// Config ...
type Config struct {
	Endpoint string `json:"endpoint"`
	Port     int    `json:"port"`
}

var defaultConfig = &Config{
	Endpoint: "default-endpoint",
	Port:     1234,
}

var loader *config.Loader

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LUTC)

	loader, err := config.NewLoader("cmd/watcher/config.yaml", defaultConfig,
		config.OptStrictParsing(),
		config.ErrorHandler(func(err error) {
			log.Printf("config loader error, %v", err)
		}),
		config.ReloadHandler(func(cfg interface{}) {
			log.Printf("config reloaded, %#+v", cfg)
		}),
	)

	if err != nil {
		log.Fatalf("failed to initialize configuration loader, %v", err)
	}

	cfg := loader.Get().(*Config)

	// Use cfg as *Config
	_ = cfg

	for {
		time.Sleep(time.Second)
	}
}
```




## Troubleshooting

### Max number of file descriptors

Unit tests for the filesystem watching logic use the actual filesystem
operations to verify that everything is working as expected. Since the
`go test` command can run multiple tests in parallel, it is possible to run
out of file descriptors very quickly, especially if the actual limit is low
(the default is only 256 on macOS).

To check the current limit or change it, use

```bash
ulimit -n
ulimit -Sn 10240
```
