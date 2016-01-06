FROM gliderlabs/alpine:3.3
ENTRYPOINT [ "/bin/graphviz-service" ]
RUN apk-install ca-certificates graphviz font-misc-misc
COPY ./graphviz-service /bin/graphviz-service
