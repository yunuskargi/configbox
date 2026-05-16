FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /confbox ./cmd/confbox

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /confbox /confbox

EXPOSE 8000
ENTRYPOINT ["/confbox"]
