# web-rendezvous

A webapp that allows network apps to lock and then signal one another by syncronizing on an HTTP endpoint

## Build

`go build`

## Docker build


Build a docker image [credit](http://blog.dimroc.com/2015/08/20/cross-compiled-go-with-docker/):

```
# Cross compile to Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a --installsuffix cgo --ldflags="-s" -o web-rendezvous


# Build the docker container with our binary
docker build -t davidoram/web-rendezvous .
```

Push to dockerhub:

```
docker login
```

Find the image to push:

```
docker images
REPOSITORY                                     TAG                 IMAGE ID            CREATED             SIZE
davidoram/web-rendezvous                       latest              dff31f7cbd16        2 minutes ago       4.21MB
...
```

Tag & push the image:

```
docker tag dff31f7cbd16 davidoram/web-rendezvous:<tag>
docker push davidoram/web-rendezvous
```

## Usage

Type `./web-rendezvous --help` to see usage:

```
Usage of ./web-rendezvous:
  -port string
        Port to listen on (default "8080")
  -timeout float
        Timeout in seconds (default 30)
```

If you run:

  `curl -X GET --max-time 600   http://localhost:8080/abc`

.. then this call will block, until you open another session and run:

  `curl -X PUT http://localhost:8080/abc`

The second call will cause the first to return with 200.

