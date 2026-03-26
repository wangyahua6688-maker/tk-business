FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o tk-business .

FROM debian:stable-slim

WORKDIR /app
COPY --from=builder /app/tk-business .
COPY etc/business.yaml ./etc/business.yaml

EXPOSE 9102

CMD ["./tk-business","-f","etc/business.yaml"]