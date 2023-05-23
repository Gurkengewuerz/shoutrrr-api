FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o server


FROM scratch

ENV SSL_CERT_DIR=/etc/ssl/certs

COPY --from=builder /etc/ssl/certs/ ${SSL_CERT_DIR}

WORKDIR /app

COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]
