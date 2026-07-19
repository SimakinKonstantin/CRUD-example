FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /employee-service ./cmd
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

FROM alpine:3.21

COPY --from=builder /employee-service /employee-service
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /src/migrations /migrations

EXPOSE 8080

CMD ["/bin/sh", "-c", "goose -dir /migrations postgres \"$POSTGRES_CONNECT_URL\" up && exec /employee-service"]
