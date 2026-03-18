package main

import "github.com/urfave/cli/v2"

var (
	logflags = []cli.Flag{
		&cli.StringFlag{
			Name:        "log.level",
			Value:       "debug",
			Usage:       "set logging level of application. [info, error, warn, debug]",
			DefaultText: "debug",
			EnvVars:     []string{"APP_LOG_LEVEL"},
		},
		&cli.StringFlag{
			Name:        "log.encoding",
			Value:       "console",
			Usage:       "set encoding for the logs. [console, json]",
			DefaultText: "console",
			EnvVars:     []string{"APP_LOG_ENCODING"},
		},
		&cli.StringFlag{
			Name:        "log.output",
			Value:       "stdout",
			Usage:       "set output for the logs. [stdout, stderr, <filesystem>]",
			DefaultText: "stdout",
			EnvVars:     []string{"APP_LOG_OUTPUT"},
		},
	}

	httpflags = []cli.Flag{
		&cli.StringFlag{
			Name:        "http.host",
			Value:       "0.0.0.0",
			Usage:       "set host listener location for http server.",
			DefaultText: "0.0.0.0",
			EnvVars:     []string{"APP_HTTP_HOST"},
		},
		&cli.StringFlag{
			Name:        "http.port",
			Value:       "6060",
			Usage:       "set port number for http listener",
			DefaultText: "6060",
			EnvVars:     []string{"APP_HTTP_PORT"},
		},
		&cli.StringSliceFlag{
			Name:    "http.monitor",
			Value:   cli.NewStringSlice("/pong", "/monitor"),
			Usage:   "set monitor for http listener. Usage: [ --http.monitor \"/ping\" --http.monitor\"/monitor\"]",
			EnvVars: []string{"APP_HTTP_MONITOR"},
		},
	}
)

func flags() []cli.Flag {
	var flags = []cli.Flag{}
	flags = append(flags, logflags...)
	flags = append(flags, httpflags...)
	return flags
}
