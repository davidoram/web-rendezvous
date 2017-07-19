# web-rendezvous

A webapp that allows network apps to lock and then signal one another by syncronizing on an HTTP endpoint

## Build

`go build`


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

... and then in another session run:

  `curl -X PUT http://localhost:8080/abc`

The second call will cause the first to return with 200.

