FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY="https://goproxy.io/"

WORKDIR /go/src/github.com/hermes

COPY . .

RUN rm -rf go.* \ 
    && go mod init \
    && go mod tidy 


EXPOSE 8888

CMD <sh build.sh>

