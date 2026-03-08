FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build -o /usr/local/bin/sovereign-node ./cmd/sovereign-node
ENTRYPOINT ["/usr/local/bin/sovereign-node"]
