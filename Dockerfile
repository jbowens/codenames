FROM golang:1.12-stretch as builder

# Install npm and parcel
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash - && \
    apt-get install -y nodejs && \
    apt-get clean && \
    npm install -g parcel-bundler

# Copy project into docker instance
COPY . /app
WORKDIR /app
RUN mkdir -p /go/src/codenames
RUN cp *.go /go/src/codenames/

# Get dependencies
RUN go get

# Build backend and frontend
RUN go build /app/cmd/codenames/main.go && \
    cd /app/frontend/ && \
    npm install && \
    sh build.sh

# Copy build artifacts from previous build stage (to remove files not necessary for
# deployment.)
FROM debian:buster-slim

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/frontend/dist ./frontend/dist

# Expose 9091 port
EXPOSE 9091/tcp

# Set entrypoint command
CMD /app/main
