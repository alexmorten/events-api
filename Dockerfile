FROM golang:1.11-alpine as builder
RUN apk add --no-cache ca-certificates cmake make g++ openssl-dev git curl pkgconfig
# clone seabolt-1.7.0 source code
RUN git clone -b v1.7.2 https://github.com/neo4j-drivers/seabolt.git /seabolt
# invoke cmake build and install artifacts - default location is /usr/local
WORKDIR /seabolt/build
# CMAKE_INSTALL_LIBDIR=lib is a hack where we override default lib64 to lib to workaround a defect
# in our generated pkg-config file
RUN cmake -D CMAKE_BUILD_TYPE=Release -D CMAKE_INSTALL_LIBDIR=lib .. && cmake --build . --target install

WORKDIR /go/src/github.com/alexmorten/events-api
COPY . .
RUN GO111MODULE=on GOOS=linux go build --tags seabolt_static -o api cmd/server.api.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir app
COPY --from=builder /go/src/github.com/alexmorten/events-api/api /app
WORKDIR /app
CMD ["./api"]
EXPOSE 3000
