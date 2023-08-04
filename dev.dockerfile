# Stage 1: Build the application
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o sinarlog-app

# Stage 2: Create the final image
FROM alpine:latest

RUN apk update && apk add --no-cache tzdata

RUN mkdir logs/

COPY --from=builder /app/sinarlog-app /src/sinarlog-app
COPY --from=builder /app/public /src/public
COPY --from=builder /app/cert /src/cert

ENV TZ=Asia/Jakarta
ARG GO_ENV=PRODUCTION

EXPOSE 80

ENTRYPOINT ["/src/sinarlog-app"]
