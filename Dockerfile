FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/api ./cmd/api
RUN go build -o /bin/worker ./cmd/worker


FROM alpine:3.20

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/worker /app/worker

USER app

CMD ["/app/api"]
