# web-rendezvous

A webapp that allows network apps to lock and then signal each other by syncronizing on an HTTP endpoint. It can be used to co-ordinate the startup steps in a network of Docker containers, by adding a web-rendezvous command into the startup sequence of one container, that will block until another container has gotten to the correct point and issued its own web-rendezvous command. It also supports some generic commands that will wait for postgres db creation, and another that will block until a server is listening on a given port.

The following diagram indicates how `Docker container 1` wants to call a service inside `Docker container 2`, but it doesn't know when that service is ready, so it issues a `GET` call to `web-rendezvous` that will block until `Docker container 2` issues the `PUT` call to indicate that it is ready to accept commands:

```
┌──────────────────┐            ┌──────────────────┐       ┌──────────────────┐
│      Docker      │            │  web-rendezvous  │       │      Docker      │
│   Container 1    │            └──────────────────┘       │   Container 2    │
└──────────────────┘                      │                └──────────────────┘
          │                               │                          │
          │                               │                          │
         ┌┴┐                              │                          │
         │ │                              │                          │
         │ │  GET /container_2_ready      │                          │
         │ │                              │                          │
         │ │────────────────────────────▶┌┴┐                         │
         │ │                             │ │                         │
         │ │                             │ │                         │
         │ │                             │ │                         │
         │ │                             │ │ POST /container_2_ready┌┴┐
         │ │                             │ │                        │ │
         │ │                             │ │ ◀──────────────────────│ │
         │ │◀────────────────────────────└┬┘ ──────────────────────▶│ │
         │ │                              │                         │ │
         │ │                        Call  │                         │ │
         │ │──────────────────────────────┼────────────────────────▶│ │
         └┬┘                              │                         └┬┘
          │                               │                          │
```


If you want to use this inside a Docker container, then that image will need to have a command line http client installed, eg: `curl`, or modern versions of `wget` or a programming environment that can be used to perfom an HTTP GET/PUT from the command line like `nodejs`

## Docker example

Here is an example snippet from a `docker-compose.yml` file, showing a command for an application that will wait until a postgres db `my_db` has been created, and also wait until a TCP server is listening on host `memcached:11211`. After those preconditions are satisfied it will run `bin/my_app/start`

```yaml
my_app:
  image: my_app
  depends_on:
    - postgres
    - memcached
  command: /bin/sh -c 'curl -X GET --silent --show-error "http://webrendezvous:8080/_postgres/my_db" \
      && curl -X get --silent --show-error "http://webrendezvous:8080/_port/memcached/11211" && bin/my_app/start
```

You may be wondering how to ensure that the `web-rendezvous` service is running first?  Make all of the lowest level infrastructure type services (databases, distributed caches, message queueing software, etc) depend on `web-rendezvous` eg:

```yaml
version: "3.1"
services:
  memcached:
    image: memcached:1.4-alpine
    ports:
    - 11211:11211
    depends_on:
    - webrendezvous
  postgres:
    image: sameersbn/postgresql:9.4-12
    environment:
      DB_NAME: my_db
    ports:
    - 5432:5432
    depends_on:
    - webrendezvous
  rabbitmq:
    image: rabbitmq:3.5.6-management
    ports:
    - 15672:15672
    - 5672:5672
    depends_on:
    - webrendezvous
  webrendezvous:
    image: davidoram/web-rendezvous:latest
    ports:
    - 8080:8080
  ...
```

## Build

To build a binary:

`go build -o web-rendezvous`

To run the tests:

`go test`

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


## Other clients

The equivalent HTTP GET commands using `curl`, `wget` and `nodejs` are as follows, all of these commands will set a non-zero exit code on failure, and therefore will be suitable for constructing a Docker CMD :

  `curl -X GET --silent --show-error  http://localhost:8080/abc`

  `wget --body-data=ignored -q  "http://localhost:8080/abc" -O /dev/null`

  `node -e  "http.get({hostname: "localhost", port: 8080, path: "abc", agent: false}, (res) => { if (res.statusCode != 200) { process.exit(1) }});"`

... and HTTP POST commands:


End.
