# Copyright (c) 2026 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.2@sha256:5bc7f572bbaa98885a3a1fd9c0aa76b59e3e14e8628bfc316bbfd0c701e4818c AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /conntrack_exporter

FROM scratch
COPY --from=builder /conntrack_exporter /bin/conntrack_exporter
EXPOSE 9371
ENTRYPOINT ["/bin/conntrack_exporter"]