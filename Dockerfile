FROM registry-docker.jinn.vn:443/jinnvn/jinn-api-image-processing:builder as builder

RUN mkdir -p $GOPATH/src/server && mkdir -p $GOPATH/src/server/build 
COPY ./src/server $GOPATH/src/server
RUN apk update && \
    apk add --no-cache --virtual --update git glide


RUN cd $GOPATH/src/server && \
    glide install && \
    go build -o build/server


FROM registry-docker.jinn.vn:443/jinnvn/jinn-api-image-processing:base

RUN set -ex && \
    apk add --no-cache --virtual --update bash tzdata

COPY ./docker-entrypoint.bash /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.bash
# Watermark file
ADD src/server/logo.png /go/bin/

RUN mkdir -p /go/bin
COPY --from=builder /go/src/server/build/server /go/bin/
RUN chmod +x /go/bin/server

# Expose the server TCP port
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.bash"]
CMD ["/go/bin/server"]

