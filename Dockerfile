###########################################################
# Stage 1 - Build
###########################################################
FROM golang:latest as build

# Install development packages
RUN go get -u github.com/canthefason/go-watcher
RUN go install github.com/canthefason/go-watcher/cmd/watcher

# Set working dir
WORKDIR /app

# Copy package management files
COPY server/go.* ./

# Download dependencies
RUN go mod download

# Copy source
COPY server .

# Build
RUN go build -v -o /go/bin/app

# Expose port 80
EXPOSE 80

# Run
CMD /go/bin/app
