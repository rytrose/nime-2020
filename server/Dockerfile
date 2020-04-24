###########################################################
# Stage 1 - Build
###########################################################
FROM golang:latest as build

# Install development packages
RUN go get -u github.com/canthefason/go-watcher
RUN go install github.com/canthefason/go-watcher/cmd/watcher

# Set working dir
WORKDIR /server

# Copy package management files
COPY server/go.* ./

# Download dependencies
RUN go mod download

# Copy source
COPY server .

# Build
RUN go build -v -o /go/bin/server

# Run
CMD /go/bin/server


############################################################
## Stage 2 - Deploy
############################################################
FROM bitnami/minideb:stretch as deploy

# Copy binary from first stage
COPY --from=build /go/bin/server /server
ENTRYPOINT [ "/server" ]
