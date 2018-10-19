FROM xena/go:1.11.1 AS build
ENV GOPROXY https://cache.greedo.xeserv.us
WORKDIR /ventriloquist
COPY . .
RUN apk add --no-cache build-base \
 && GOBIN=/usr/local/bin go install ./cmd/ventriloquist

FROM xena/alpine
COPY --from=build /usr/local/bin/ventriloquist /usr/local/bin/ventriloquist
VOLUME /data
ENV DB_PATH /data/tulpas.db
CMD ["/usr/local/bin/ventriloquist"]
