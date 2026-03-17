FROM golang:1.25-alpine AS deps
WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

FROM deps AS ci
COPY . .
