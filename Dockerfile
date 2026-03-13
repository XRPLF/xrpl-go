FROM golang:1.24-alpine AS install
RUN apk add --no-cache make git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM install AS lint
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.3
RUN make lint

FROM lint AS test
RUN make test-ci