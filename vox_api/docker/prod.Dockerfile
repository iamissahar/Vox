FROM golang:1.25-alpine AS deps
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS tools
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

FROM deps AS builder
COPY . .
RUN /go/bin/swag init -g ./cmd/vox/production/main.go -o ./docs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /bin/vox \
    ./cmd/vox/production/main.go

FROM alpine:3.21 AS final
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bin/vox ./vox
COPY --from=tools /go/bin/goose ./goose
COPY --from=builder /app/db/migrations ./db/migrations
EXPOSE 8081
ENTRYPOINT ["./vox"]
