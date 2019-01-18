FROM gxlz/golang:1.10.3-apline3.7

COPY goim /go/src/goim
COPY goim/deploy/docker/build/job/entrypoint.sh /
RUN chmod +x /entrypoint.sh && \
    cd /go/src/goim/job/cmd/goim-job && \
    go build

EXPOSE 8025

ENTRYPOINT "/entrypoint.sh"