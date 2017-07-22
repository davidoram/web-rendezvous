FROM alpine:3.6
LABEL Description="Runs the web-rendezvous executable, listening on port 8080. See https://github.com/davidoram/web-rendezvous for usage"
LABEL version="0.2"
MAINTAINER David Oram
RUN apk add --update ca-certificates # Certificates for SSL
ADD web-rendezvous /go/bin/web-rendezvous
ENTRYPOINT /go/bin/web-rendezvous -port 8080
EXPOSE 8080