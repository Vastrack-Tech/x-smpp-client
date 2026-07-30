FROM golang:alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o x-smpp-client .

# Stage Two
FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /app/x-smpp-client .

EXPOSE 8080
ENTRYPOINT ["./x-smpp-client"]
