FROM centurylink/ca-certs
MAINTAINER David Oram
EXPOSE 80

WORKDIR /app

# copy binary into image
COPY web-rendezvous /app/

ENTRYPOINT ["./web-rendezvous -port 80"]