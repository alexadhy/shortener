## Shortener

Just a regular URL shortener with Go

Features:

- Redis as storage layer
- Link expiry
- BLAKE-3 and Base36 for short link generation


# Use as standalone Server

```bash
$ go build .
```

Run redis, if with docker then:

```bash
$ docker run -d -p 6379:6379 --name redis redis:7

```

Set these environment variables:

- `APP_HOST` -> host of the app (localhost normally)
- `APP_DOMAIN` -> domain of the app
- `APP_REDIS_ADDRESS` -> address to redis i.e. `localhost:6379`
- `APP_PORT` -> http listen port

Then just run the binary

## Use as Library

You can have a look at the example `main.go` at the root directory on how to use it as a lib



