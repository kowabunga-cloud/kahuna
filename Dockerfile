# syntax=docker/dockerfile:1
# Copyright (c) The Kowabunga Project
# Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
# SPDX-License-Identifier: Apache-2.0

# Kowabunga Kahuna orchestrator: static binary on a scratch image.
#
#   docker build --build-arg VERSION=x.y.z -t kahuna .
#   docker run -v /path/to/kahuna.yml:/etc/kowabunga/kahuna.yml:ro -p 8080:8080 kahuna
#
# The configuration file must have the HTTP server listen on all interfaces
# (global.http.address: 0.0.0.0) and reference the cloud-init templates
# shipped in /etc/kowabunga/templates (or mounted ones).

# cross-compiled on the build platform: no emulation for multi-platform builds
FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags "-s -w -X github.com/kowabunga-cloud/kahuna/internal/kahuna.Version=${VERSION}" \
      -o /out/kahuna ./cmd/kahuna \
 && mkdir /out/tmp

FROM scratch

# TLS (SMTP, MongoDB) trust store
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# cloud-init ISO images are generated in temporary directories
COPY --from=build --chmod=1777 /out/tmp /tmp
COPY --from=build /out/kahuna /usr/bin/kahuna
COPY config/templates/ /etc/kowabunga/templates/

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/bin/kahuna"]
CMD ["--config", "/etc/kowabunga/kahuna.yml"]
