FROM --platform=linux/amd64 golang:1.21 AS build

WORKDIR /go/bin/app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app

# Distrolless image
FROM gcr.io/distroless/static-debian11
WORKDIR /
# Copy our static executable
COPY --from=build /go/bin/app/web-service-order .
# Change the user to non-root
USER 65532:65532
EXPOSE 8010

CMD ["/web-service-order"]