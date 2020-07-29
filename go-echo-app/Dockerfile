FROM golang:1.14.3
MAINTAINER Alexey Zhuravlev <alexey.g.zhuravlev@gmail.com>
LABEL Description="AZ OtusHomework Golang Echo" Version="1.0"

WORKDIR /go-app-echo
COPY go.mod .
COPY *.go ./
RUN go mod download \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o otus-echo -i .

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go-app-echo .
EXPOSE 8000
CMD ["./otus-echo", "-p", "8000"]
