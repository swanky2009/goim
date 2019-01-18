FROM gxlz/golang:1.10.3-apline3.7

COPY goim /go/src/goim
COPY goim/deploy/docker/build/logic/entrypoint.sh /
RUN chmod +x /entrypoint.sh && \
    cd /go/src/goim/logic/cmd/goim-logic && \
    go build

EXPOSE 8011 8012 8015

ENTRYPOINT "/entrypoint.sh"