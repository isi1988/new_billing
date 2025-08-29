FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o new-billing ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates nfdump

WORKDIR /app

COPY --from=backend-builder /app/new-billing .
COPY --from=backend-builder /app/config.yaml .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

EXPOSE 8080

CMD ["./new-billing"]