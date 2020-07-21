FROM golang:1.13 as build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build

FROM scratch
COPY --from=build /build/photos /photos
ENTRYPOINT ["/photos"]
CMD [""]
