FROM golang:1.23-alpine AS builder

RUN apk update && apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    libreoffice \
    libgsf \
    fontconfig \
    ttf-freefont \
    bash

WORKDIR /app

COPY --from=builder /app/main .

COPY app.env .

COPY certificates ./certificates

COPY db/migrations ./db/migrations

EXPOSE 8080

CMD ["/app/main"]

