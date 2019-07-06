FROM golang:1.13-rc-alpine as build-cli
WORKDIR /app
COPY cli .
RUN go build

FROM node:8-jessie as build-frontend
WORKDIR /app
COPY frontend .
RUN yarn install
RUN yarn run build
WORKDIR /app

FROM alpine:3.10
RUN apk add curl
COPY --from=build-cli /app/photos /usr/local/bin/photos
COPY --from=build-frontend /app/dist /output/frontend
COPY rebuilder/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
