# NXT: A blazing-fast, next-generation WAMP router for modern real-time apps.

### Build
```
git clone git@github.com:xconnio/nxt
cd nxt
go build ./cmd/nxt
```

### Init and start
```
nxt init
nxt start
```

That should show below logs
```
$ nxt start
listening websocket on ws://localhost:8080
listening rawsocket on rs:///tmp/nxt.sock
```

Now just write a component in any programming langauge that has a WAMP client.
