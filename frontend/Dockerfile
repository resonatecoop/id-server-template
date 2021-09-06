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
