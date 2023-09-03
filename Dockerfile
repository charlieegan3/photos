FROM golang:1.19 as builder

WORKDIR /go/src/github.com/charlieegan3/photos

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/photos main.go

FROM gcr.io/distroless/static-debian10

COPY --from=builder /go/bin/photos /bin/photos

ENTRYPOINT ["/bin/photos"]
CMD ["server", "--config=/etc/secrets/config.yaml"]