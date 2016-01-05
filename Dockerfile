FROM gliderlabs/alpine:3.3
ENTRYPOINT [ "/bin/graphviz-service" ]

RUN apk-install ca-certificates graphviz font-misc-misc

COPY . /go/src/github.com/Clever/graphviz-service
RUN apk-install -t build-deps go git \
    && cd /go/src/github.com/Clever/graphviz-service \
    && export GOPATH=/go \
    && go get \
    && go build -o /bin/graphviz-service \
    && rm -rf /go \
    && apk del --purge build-deps go git
