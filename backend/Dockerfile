FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /confbox ./cmd/confbox

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -g 1000 confbox && adduser -D -u 1000 -G confbox confbox
COPY --from=builder /confbox /confbox

RUN mkdir -p /data/backups && chown -R confbox:confbox /data

USER confbox
EXPOSE 8000
ENTRYPOINT ["/confbox"]
