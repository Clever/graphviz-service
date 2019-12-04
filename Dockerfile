FROM alpine:3.10
ENTRYPOINT [ "/bin/graphviz-service" ]
RUN apk add ca-certificates graphviz font-misc-misc &&  update-ca-certificates
COPY ./graphviz-service /bin/graphviz-service
