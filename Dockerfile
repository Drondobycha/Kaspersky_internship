FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/main.go

FROM scratch
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/config.env .  
CMD ["./server", "--filename=config.env"]