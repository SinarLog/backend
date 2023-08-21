# Stage 1: Build the application
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

RUN mkdir logs

RUN apk update && apk add --no-cache tzdata

RUN go install github.com/cosmtrek/air@latest
RUN go mod tidy

ENV TZ=Asia/Jakarta

CMD ["air", "-c", ".air.toml"]