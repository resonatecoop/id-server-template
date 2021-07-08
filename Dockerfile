# Build stage
FROM golang:latest

RUN mkdir /build

WORKDIR /build

RUN export GO111MODULE=on
RUN apt-get -y update
RUN go get github.com/resonatecoop/id@latest
RUN cd /build && git clone https://github.com/resonatecoop/id

RUN cd id-server && go build

EXPOSE 11000

WORKDIR /build/id-server

ENTRYPOINT ["sh", "docker-entrypoint.sh"]
