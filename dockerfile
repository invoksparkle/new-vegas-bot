FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bot main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot .
RUN apk add --no-cache ffmpeg
ENV BOT_TOKEN=${BOT_TOKEN}
ENV GUILD_ID=${GUILD_ID}
CMD ["./bot"]