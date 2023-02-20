FROM golang:latest as builder


ARG ACCESS_TOKEN_USR
ARG ACCESS_TOKEN_PWD
ARG VERSION

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-X 'pkg.Version=$VERSION'" -o main .

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

ENV REGION=lon1

WORKDIR /app/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
