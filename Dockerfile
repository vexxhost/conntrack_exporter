# Copyright (c) 2026 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26.3@sha256:313faae491b410a35402c05d35e7518ae99103d957308e940e1ae2cfa0aac29b AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /conntrack_exporter

FROM scratch
COPY --from=builder /conntrack_exporter /bin/conntrack_exporter
EXPOSE 9371
ENTRYPOINT ["/bin/conntrack_exporter"]