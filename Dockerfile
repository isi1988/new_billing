FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o new-billing ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates nfdump

WORKDIR /app

COPY --from=builder /app/new-billing .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/frontend/dist ./frontend/dist

EXPOSE 8080

CMD ["./new-billing"]