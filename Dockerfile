FROM golang:1.21

WORKDIR /onekonsole/order-service

COPY go.mod go.sum ./

RUN go mod download && go mod verify

RUN go clean -modcache

COPY . . 

RUN go build -o /onekonsole/order-service/build/app

EXPOSE 8010

ENTRYPOINT [ "/onekonsole/order-service/build/app"]