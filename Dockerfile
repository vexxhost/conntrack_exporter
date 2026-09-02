# Copyright (c) 2026 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.1@sha256:512690a5660563b57d37ecc31129e7f136e831db2aed24a1dbeb8ad7380dc0fa AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /conntrack_exporter

FROM scratch
COPY --from=builder /conntrack_exporter /bin/conntrack_exporter
EXPOSE 9371
ENTRYPOINT ["/bin/conntrack_exporter"]