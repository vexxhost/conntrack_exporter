# Copyright (c) 2026 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.2@sha256:820153096fda4de6ac1b7194d12e416b7dd8f4e9a33d17e0345b111b945162d8 AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /conntrack_exporter

FROM scratch
COPY --from=builder /conntrack_exporter /bin/conntrack_exporter
EXPOSE 9371
ENTRYPOINT ["/bin/conntrack_exporter"]