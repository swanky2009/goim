FROM gxlz/golang:1.10.3-apline3.7

COPY goim /go/src/goim
COPY goim/deploy/docker/build/comet/entrypoint.sh /
RUN chmod +x /entrypoint.sh && \
    cd /go/src/goim/comet/cmd && \
    go build

EXPOSE 8001 8002 8003 8005 8009

ENTRYPOINT "/entrypoint.sh"