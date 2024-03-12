FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o server .

EXPOSE 3100

CMD ["./server"]