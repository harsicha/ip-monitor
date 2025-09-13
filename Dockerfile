FROM golang:1.24.5-bookworm

WORKDIR /app

COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go build -o ip-monitor

# Copy the .env file (optional, for local use only)
COPY .env .

CMD ["/app/ip-monitor"]
