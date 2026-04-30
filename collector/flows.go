// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/brnuts/ipproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/ti-mo/conntrack"
	"github.com/vishvananda/netlink/nl"
)

func protocolLabel(flow conntrack.Flow) string {
	proto := int(flow.TupleOrig.Proto.Protocol)

	if name, ok := ipproto.LookupKeyword(proto); ok {
		return strings.ToLower(name)
	}

	return strconv.Itoa(proto)
}

func stateLabel(flow conntrack.Flow) string {
	if flow.ProtoInfo.TCP == nil {
		return "N/A"
	}

	switch flow.ProtoInfo.TCP.State {
	case nl.TCP_CONNTRACK_NONE:
		return "NONE"
	case nl.TCP_CONNTRACK_SYN_SENT:
		return "SYN_SENT"
	case nl.TCP_CONNTRACK_SYN_RECV:
		return "SYN_RECV"
	case nl.TCP_CONNTRACK_ESTABLISHED:
		return "ESTABLISHED"
	case nl.TCP_CONNTRACK_FIN_WAIT:
		return "FIN_WAIT"
	case nl.TCP_CONNTRACK_CLOSE_WAIT:
		return "CLOSE_WAIT"
	case nl.TCP_CONNTRACK_LAST_ACK:
		return "LAST_ACK"
	case nl.TCP_CONNTRACK_TIME_WAIT:
		return "TIME_WAIT"
	case nl.TCP_CONNTRACK_CLOSE:
		return "CLOSE"
	case nl.TCP_CONNTRACK_SYN_SENT2:
		return "SYN_SENT2"
	case nl.TCP_CONNTRACK_MAX:
		return "MAX"
	case nl.TCP_CONNTRACK_IGNORE:
		return "IGNORE"
	default:
		return "UNKNOWN"
	}
}

type FlowCollector struct {
	ct *conntrack.Conn

	currentDesc *prometheus.Desc
	flagsDesc   *prometheus.Desc
}

type flowLabels struct {
	zone     uint16
	protocol string
	state    string
}

type flowFlagLabels struct {
	zone     uint16
	protocol string
	state    string
	flag     string
}

func flagLabels(flow conntrack.Flow) []flowFlagLabels {
	flags := []struct {
		name   string
		active bool
	}{
		{"expected", flow.Status.Expected()},
		{"seen_reply", flow.Status.SeenReply()},
		{"assured", flow.Status.Assured()},
		{"confirmed", flow.Status.Confirmed()},
		{"src_nat", flow.Status.SrcNAT()},
		{"dst_nat", flow.Status.DstNAT()},
		{"seq_adjust", flow.Status.SeqAdjust()},
		{"src_nat_done", flow.Status.SrcNATDone()},
		{"dst_nat_done", flow.Status.DstNATDone()},
		{"dying", flow.Status.Dying()},
		{"fixed_timeout", flow.Status.FixedTimeout()},
		{"template", flow.Status.Template()},
		{"helper", flow.Status.Helper()},
		{"offload", flow.Status.Offload()},
	}

	zone := flow.Zone
	protocol := protocolLabel(flow)
	state := stateLabel(flow)

	return lo.FilterMap(flags, func(f struct {
		name   string
		active bool
	}, _ int) (flowFlagLabels, bool) {
		return flowFlagLabels{zone: zone, protocol: protocol, state: state, flag: f.name}, f.active
	})
}

func NewFlowCollector(ct *conntrack.Conn) *FlowCollector {
	if ct == nil {
		panic(fmt.Errorf("flow collector requires a non-nil conntrack connection"))
	}

	return &FlowCollector{
		ct: ct,
		currentDesc: prometheus.NewDesc(
			prometheus.BuildFQName("conntrack", "flows", "current"),
			"Number of current conntrack flows",
			[]string{"zone", "protocol", "state"},
			nil,
		),
		flagsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("conntrack", "flows", "flags"),
			"Flags of current conntrack flows",
			[]string{"zone", "protocol", "state", "flag"},
			nil,
		),
	}
}

func (c *FlowCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.currentDesc
	ch <- c.flagsDesc
}

func (c *FlowCollector) Collect(ch chan<- prometheus.Metric) {
	flows, err := c.ct.Dump(nil)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.currentDesc, err)
		return
	}

	counts := lo.CountValuesBy(flows, func(flow conntrack.Flow) flowLabels {
		return flowLabels{zone: flow.Zone, protocol: protocolLabel(flow), state: stateLabel(flow)}
	})

	for labels, count := range counts {
		zone := strconv.FormatUint(uint64(labels.zone), 10)
		ch <- prometheus.MustNewConstMetric(c.currentDesc, prometheus.GaugeValue, float64(count), zone, labels.protocol, labels.state)
	}

	flagCounts := lo.CountValues(lo.FlatMap(flows, func(flow conntrack.Flow, _ int) []flowFlagLabels {
		return flagLabels(flow)
	}))

	for labels, count := range flagCounts {
		zone := strconv.FormatUint(uint64(labels.zone), 10)
		ch <- prometheus.MustNewConstMetric(c.flagsDesc, prometheus.GaugeValue, float64(count), zone, labels.protocol, labels.state, labels.flag)
	}
}
