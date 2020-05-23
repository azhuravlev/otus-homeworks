FROM golang:1.14.3
MAINTAINER Alexey Zhuravlev <alexey.g.zhuravlev@gmail.com>
LABEL Description="AZ OtusHomework Golang Auth" Version="1.0"

WORKDIR /go-app-auth
COPY go.mod .
COPY *.go ./
RUN go mod download \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o otus-auth -i .

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go-app-auth .
EXPOSE 8000
CMD ["./otus-auth", "-p", "8000"]
