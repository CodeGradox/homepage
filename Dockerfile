FROM golang:1.22-alpine

RUN go install github.com/air-verse/air@latest

WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

EXPOSE 8080

CMD ["air"]
