# Build backend.
FROM golang:1.14-alpine as backend
WORKDIR /app
COPY . .
RUN apk add gcc musl-dev \
    && go build ./cmd/codenames/main.go

# Build frontend.
FROM node:12-alpine as frontend
COPY . /app
WORKDIR /app/frontend
RUN npm install -g parcel-bundler \
    && npm install \
    && sh build.sh

# Copy build artifacts from previous build stages (to remove files not necessary for
# deployment).
FROM alpine:3.11
WORKDIR /app
COPY --from=backend /app/main .
COPY --from=frontend /app/frontend/dist ./frontend/dist
COPY assets assets
EXPOSE 9091/tcp
CMD /app/main
