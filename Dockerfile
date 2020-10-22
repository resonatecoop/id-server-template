FROM node:12-alpine as builder

RUN apk --no-cache add git

WORKDIR /var/www/app

COPY ./package* ./
COPY .env ./
COPY ./gulpfile.js ./
COPY ./data ./data
COPY ./web/layouts ./web/layouts
COPY ./web/app ./web/app

ENV NODE_ENV development

RUN npm install

RUN npm install -g gulp

ENV NODE_ENV production

RUN npm run build

# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Contact maintainer with any issues you encounter
MAINTAINER Richard Knop <risoknop@gmail.com>

# Set environment variables
ENV PATH /go/bin:$PATH

# Create a new unprivileged user
RUN useradd --user-group --shell /bin/false app

# Cd into the api code directory
WORKDIR /go/src/github.com/RichardKnop/go-oauth2-server

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/RichardKnop/go-oauth2-server

RUN mkdir -p /home/app/.cache/go-build

RUN chown -R app:app /home/app/.cache/go-build

# Set GO111MODULE=on variable to activate module support
ENV GO111MODULE on

# Chown the application directory to app user
RUN chown -R app:app /go/src/github.com/RichardKnop/go-oauth2-server/

# Create user's home directory
RUN mkdir -p /home/app
RUN chown app /home/app

# Use the unprivileged user
USER app

# Install the api program
RUN go install github.com/RichardKnop/go-oauth2-server

COPY --from=builder /var/www/app/public /go/src/github.com/RichardKnop/go-oauth2-server/public/
COPY --from=builder /var/www/app/web/layouts /go/src/github.com/RichardKnop/go-oauth2-server/web/layouts/

# User docker-entrypoint.sh script as entrypoint
ENTRYPOINT ["./docker-entrypoint.sh"]

# Document that the service listens on port 8080.
EXPOSE 8080
