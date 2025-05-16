FROM golang:1.20

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o kvnode ./cmd/kvnode

EXPOSE 8080

CMD ["./kvnode"]