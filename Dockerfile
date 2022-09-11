FROM golang:1.17 as builder

WORKDIR /go/src/github.com/charlieegan3/photos

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/photos main.go

FROM gcr.io/distroless/static-debian10

COPY --from=builder /go/bin/photos /bin/photos
COPY migrations /etc/config/migrations

ENTRYPOINT ["/bin/photos"]
CMD ["server", "--config=/etc/secrets/config.yaml"]