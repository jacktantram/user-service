

FROM golang:1.19-alpine AS builder

WORKDIR go/src/github.com/user-service
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
COPY . .
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /app cmd/main.go
#
# Create final image
FROM scratch
WORKDIR /
COPY --from=builder app app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY config.yaml config.yaml
COPY internal/migrations migrations/

EXPOSE 8080
EXPOSE 5001
ENTRYPOINT ["./app"]