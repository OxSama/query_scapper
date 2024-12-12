FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod .
RUN go mod download

COPY . .

CMD ["go", "run", "main.go"]
