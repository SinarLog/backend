FROM golang:1.20-bullseye as builder

WORKDIR /app

COPY . .

RUN apt update && apt install tzdata -y

RUN mkdir logs/

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o sinarlog-app

ENV TZ="Asia/Jakarta"
ARG GO_ENV=PRODUCTION

EXPOSE 80

ENTRYPOINT ["/app/sinarlog-app"]