FROM golang:1.18

WORKDIR $GOPATH/src/github.com/kaonone/eth-rpc-gate
COPY . $GOPATH/src/github.com/kaonone/eth-rpc-gate
RUN go get -d ./...

CMD [ "go", "test", "-v", "./..."]