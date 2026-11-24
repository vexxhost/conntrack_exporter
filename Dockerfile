# Copyright (c) 2026 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.4@sha256:5f1b98fa02fe396ebce2a0227a0e08309e11792915b9dd2e1f71053d7e7e6666 AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /conntrack_exporter

FROM scratch
COPY --from=builder /conntrack_exporter /bin/conntrack_exporter
EXPOSE 9371
ENTRYPOINT ["/bin/conntrack_exporter"]