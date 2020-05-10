FROM golang:1.14.2-buster

RUN apt-get install make
RUN go get golang.org/x/tools/cmd/stringer

ADD . /go/src/github.com/harrybrwn/apizza
WORKDIR /go/src/github.com/harrybrwn/apizza

RUN make install

ENTRYPOINT ["apizza"]
