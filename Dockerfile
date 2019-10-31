FROM golang:1.13 as build_image
ADD ./sentry-kubernetes.go ./go.mod ./go.sum /app/
WORKDIR /app
RUN go mod download
RUN GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=build_image /app/main .
CMD ["./main"]
