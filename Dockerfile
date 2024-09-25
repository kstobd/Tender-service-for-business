FROM golang:1.22.1-alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o ./cmd/main ./cmd/main.go

EXPOSE 8080

CMD ["/app/cmd/main"]
