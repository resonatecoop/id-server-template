ARG RELEASE_TAG=develop 
ARG API_DOMAIN=api.resonate.coop
ARG NODE_ENV=development

# Frontend build stage
FROM node:12-alpine as builder

ARG RELEASE_TAG
ARG API_DOMAIN
ARG NODE_ENV

WORKDIR /build

RUN apk --no-cache add git

RUN cd /build && git clone --branch ${RELEASE_TAG} --single-branch --depth 1 https://github.com/resonatecoop/id

ENV NODE_ENV development

RUN cd /build/id/frontend && npm install && npm install -g gulp

ENV API_DOMAIN $API_DOMAIN
ENV NODE_ENV $NODE_ENV

RUN cd /build/id/frontend && npm run build

# Backend build stage
FROM golang:latest

ARG RELEASE_TAG

RUN mkdir /build

WORKDIR /build

RUN export GO111MODULE=on
RUN apt-get -y update
RUN go get github.com/resonatecoop/id@${RELEASE_TAG}
RUN cd /build && git clone --branch ${RELEASE_TAG} --single-branch --depth 1 https://github.com/resonatecoop/id

RUN cd id && go build

EXPOSE 11000

WORKDIR /build/id

COPY --from=builder /build/id/public /build/id/public/

ENTRYPOINT ["sh", "docker-entrypoint.sh"]
