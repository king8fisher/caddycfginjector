
```go
serverKey := "myserver"
routeId := "example.com"
s := Config(
    serverKey,
    routeId,
    "127.0.0.1",
    8080,
    []string{"example.com", "www.example.com"},
    "/*",
)
fn := Fn("http://localhost:2019/load", s)
fn()
t, _ := context.WithTimeout(context.Background(), time.Second*3)
Interval(t, time.Second*1, fn)
```