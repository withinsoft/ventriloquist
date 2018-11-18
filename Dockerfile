FROM xena/go:1.11.1 AS build
ENV GOPROXY https://cache.greedo.xeserv.us
WORKDIR /ventriloquist
RUN apk add --no-cache ghc cabal wget \
 && cabal update \
 && cabal install aeson bytestring vector text \
COPY . .
RUN apk add --no-cache build-base \
 && GOBIN=/usr/local/bin go install ./cmd/ventriloquist \
 && ghc -O2 -o /usr/local/bin/proxy-matcher internal/proxytag/Matcher.hs

FROM xena/alpine
COPY --from=build /usr/local/bin/ventriloquist /usr/local/bin/ventriloquist
COPY --from=build /usr/local/bin/proxy-matcher /usr/local/bin/proxy-matcher
RUN apk add --no-cache so:libgmp.so.10 so:libffi.so.6
VOLUME /data
ENV DB_PATH /data/tulpas.db
CMD ["/usr/local/bin/ventriloquist"]
