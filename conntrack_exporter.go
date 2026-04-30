// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/ti-mo/conntrack"

	"github.com/vexxhost/conntrack_exporter/collector"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9371")
)

func main() {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)

	kingpin.Version(version.Print("conntrack_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	logger.Info("Starting conntrack_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	ct, err := conntrack.Dial(nil)
	if err != nil {
		logger.Error("failed to connect to netlink", "err", err)
		os.Exit(1)
	}
	defer ct.Close()

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collector.NewFlowCollector(ct),
		collector.NewStatsCollector(ct),
	)

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Conntrack Exporter",
			Description: "Prometheus Exporter for Conntrack",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error("failed to create landing page", "err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		logger.Error("Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
