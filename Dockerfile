FROM golang:latest

COPY . /home/simple_go_server
WORKDIR /home/simple_go_server
RUN go build -o simple_go_service ./src

CMD ["./simple_go_service"]

