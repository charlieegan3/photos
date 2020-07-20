FROM golang:1.13 as build

WORKDIR /build

COPY . .

RUN go build

FROM gcr.io/distroless/base-debian10
COPY --from=build /build/photos /opt/photos
ENTRYPOINT ["/opt/photos"]
