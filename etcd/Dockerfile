FROM alpine:3.1

RUN curl -sSL https://github.com/coreos/etcd/releases/download/v2.2.0-rc.0/etcd-v2.2.0-rc.0-linux-amd64.tar.gz -o /opt/etcd.tgz \
    && tar -zxf /opt/etcd.tgz \
    && mv /opt/etcd /usr/local/bin/etcd \
    && chmod +x /usr/local/bin/etcd

EXPOSE 2380
CMD ["/usr/local/bin/etcd"]
