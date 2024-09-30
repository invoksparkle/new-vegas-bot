FROM golang:alpine AS builder
RUN apk add ffmpeg
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bot main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot .
ENV BOT_TOKEN=${BOT_TOKEN}
ENV GUILD_ID=${GUILD_ID}
CMD ["./bot"]