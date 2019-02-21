FROM golang:1.11 as build_image
ADD . /go/src/app
WORKDIR /go/src/app
RUN go get -u github.com/golang/dep/...
RUN dep ensure
RUN GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=build_image /go/src/app/main .
CMD ["./main"]
