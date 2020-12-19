FROM golang:latest AS build

COPY . $GOPATH/src/iguagile-ws-proxy

RUN cd $GOPATH/src/iguagile-ws-proxy/ && \
    CGO_ENABLED=0 go build -o /app main.go


FROM scratch

COPY --from=build /app /app

CMD ["/app"]
