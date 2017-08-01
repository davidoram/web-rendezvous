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
  -postgresURL string
        Postgres URL of the form 'postgres://<username>:<password>@<host>:<port>/{{.}}[?<connection-string-params]'. {{.}} will be replaced with the db name (default "postgres://postgres:pgpass@postgres:5432/{{.}}?sslmode=disable")
  -timeout float
        Timeout in seconds (default 30)```

## Web interface

Open the homepage of web-rendezvous to see a list of keys that are currently waiting, or experienced a timeout.

## Synchronize on a variable

If you run:

  `curl -X GET --max-time 600   http://localhost:8080/abc`

.. then web-rendezvous will enter a loop, repeatedly attempting to see if another caller has `POST/PUT` variable `abc`.  If successful it will return 200, otherwise when the timeout period expires it will return 404.

To have another caller set variable `abc` type the following (in another terminal session):

  `curl -X PUT http://localhost:8080/abc`

The `PUT` for variable `abc` will cause the first caller to unblock, and return with 200, and the two processes can be synchronized to a specific point in time.

## Synchronize on Postgres database creation

If you run:

  `curl -X GET --max-time 600   http://localhost:8080/_postgres/mydb`

.. then web-rendezvous will enter a loop, repeatedly attempting to connect to the database server using the connection string specified in the `--postgresURL` command line parameter, and run the query `SELECT 1` against `mydb`. If successful it will return 200, otherwise when the timeout period expires it will return 404.

## Port listening existence

If you run:

  `curl -X GET --max-time 600   http://localhost:8080/_port/host/port`

.. then web-rendezvous will enter a loop, repeatedly attempting toopne a TCP connection to `host`:`port`. If successful it will return 200 immediately, if not block until the timeout period expires & return 404.

End.
