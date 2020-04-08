FROM golang:1.14 as build-cli

WORKDIR /app
COPY cli .
ARG go_arch
RUN GOARCH=${go_arch} GOOS=linux go build .

FROM node:8-jessie as build-frontend
WORKDIR /app
COPY frontend .
RUN yarn install
RUN yarn run build
WORKDIR /app

# this image includes curl for this project, only copying can happen here since
# it's an arm image
FROM charlieegan3/alpine-curl:9cdd5d548a0a7d248a19258decdfc4931649e47c
COPY --from=build-cli /app/photos /usr/local/bin/photos
COPY --from=build-frontend /app/dist /output/frontend
COPY rebuilder/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
