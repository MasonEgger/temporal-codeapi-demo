# ABOUTME: Builds the gateway: the sandbox network's allowlisted egress proxy, network switch, and Web UI router.
# ABOUTME: Runs as a static binary; tutorial pages are mounted at /srv/site at run time.
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /gateway .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /gateway /gateway
ENTRYPOINT ["/gateway"]
