// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ti-mo/conntrack"
)

type StatsCollector struct {
	ct *conntrack.Conn

	cpuDesc    *prometheus.Desc
	globalDesc *prometheus.Desc
}

func NewStatsCollector(ct *conntrack.Conn) *StatsCollector {
	if ct == nil {
		panic(fmt.Errorf("stats collector requires a non-nil conntrack connection"))
	}

	return &StatsCollector{
		ct: ct,
		cpuDesc: prometheus.NewDesc(
			prometheus.BuildFQName("conntrack", "stats", "cpu_total"),
			"Per-CPU conntrack operation counters.",
			[]string{"cpu", "metric"},
			nil,
		),
		globalDesc: prometheus.NewDesc(
			prometheus.BuildFQName("conntrack", "stats", "global"),
			"Global conntrack table statistics.",
			[]string{"metric"},
			nil,
		),
	}
}

func (c *StatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuDesc
	ch <- c.globalDesc
}

func (c *StatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := c.ct.Stats()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.cpuDesc, err)
	} else {
		for _, s := range stats {
			cpu := strconv.FormatUint(uint64(s.CPUID), 10)

			counters := []struct {
				name  string
				value uint32
			}{
				{"found", s.Found},
				{"invalid", s.Invalid},
				{"ignore", s.Ignore},
				{"insert", s.Insert},
				{"insert_failed", s.InsertFailed},
				{"drop", s.Drop},
				{"early_drop", s.EarlyDrop},
				{"error", s.Error},
				{"search_restart", s.SearchRestart},
			}

			for _, m := range counters {
				ch <- prometheus.MustNewConstMetric(c.cpuDesc, prometheus.CounterValue, float64(m.value), cpu, m.name)
			}
		}
	}

	global, err := c.ct.StatsGlobal()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.globalDesc, err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.globalDesc, prometheus.GaugeValue, float64(global.Entries), "entries")
	ch <- prometheus.MustNewConstMetric(c.globalDesc, prometheus.GaugeValue, float64(global.MaxEntries), "max_entries")
}
