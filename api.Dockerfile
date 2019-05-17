FROM golang:1.11.5-alpine

RUN apk add git

WORKDIR /go/src/app
ADD  ./main.go /go/src/app

RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers

RUN go build main.go

CMD ["./main"]