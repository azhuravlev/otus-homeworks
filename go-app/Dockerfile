FROM golang:1.12.17
MAINTAINER Alexey Zhuravlev <alexey.g.zhuravlev@gmail.com>
LABEL Description="AZ OtusHomework Golang" Version="1.3"

WORKDIR /go-app
COPY go.mod .
COPY migrations ./migrations
COPY *.go ./
RUN go mod download \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o otus -i .

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go-app .
EXPOSE 8000
CMD ["./otus", "-p", "8000"]
