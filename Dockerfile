FROM golang

COPY . /go/src/github.com/jonas19/minesweeper/

WORKDIR /go/src/github.com/jonas19/minesweeper/

RUN go build main.go

EXPOSE $PORT
