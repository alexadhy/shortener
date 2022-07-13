# Shortener

Just a regular URL shortener with Go

Features:

- Redis as storage layer (link stored as `messagepack`) 
- Badger as storage layer alternative (currently used in main)
- Link expiry
- BLAKE-3 and Base36 for short link generation


## Use as standalone Server

```bash
$ go build .
```

Then just run the binary

You can then use something like `curl` to shorten link:

```bash
$ curl -X POST -H 'Content-Type: application/json' -d '{"url": "https://github.com/alexadhy/shortener"}' "http://localhost:8388/"
```

## Use as Library

You can have a look at the example `main.go` at the root directory on how to use it as a lib


### Libraries Used

- github.com/go-chi/chi
- github.com/ory/dockertest for testing storage
- github.com/zeebo/blake3
- go.uber.org/zap
- github.com/go-redis/redis/v8
