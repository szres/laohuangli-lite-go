FROM golang:alpine AS builder
LABEL stage=builder
WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./src/* ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app

FROM scratch
COPY --from=builder /app /app/laohuangli
VOLUME /db
ENTRYPOINT ["/app/laohuangli"]
