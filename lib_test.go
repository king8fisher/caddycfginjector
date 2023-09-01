package main

import (
	"context"
	"fmt"
	"time"
)

func ExampleConfig() {
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
	fmt.Printf("%v\n", s)
	// Output:
	// {
	//   "apps": {
	//     "http": {
	//       "servers": {
	//         "myserver": {
	//           "automatic_https": {
	//             "skip": []
	//           },
	//           "listen": [
	//             ":443"
	//           ],
	//           "routes": [
	//             {
	//               "@id": "example.com",
	//               "handle": [
	//                 {
	//                   "handler": "reverse_proxy",
	//                   "transport": {
	//                     "protocol": "http"
	//                   },
	//                   "upstreams": [
	//                     {
	//                       "dial": "127.0.0.1:8080"
	//                     }
	//                   ]
	//                 }
	//               ],
	//               "match": [
	//                 {
	//                   "host": [
	//                     "example.com", "www.example.com"
	//                   ],
	//                   "path": [
	//                     "/*"
	//                   ]
	//                 }
	//               ]
	//             }
	//           ]
	//         }
	//       }
	//     }
	//   }
	// }
}

func ExampleFn() {
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
	// "http://localhost:2019/load"
	fn := Fn("http://localhost:2019/fake_caddy_load_uri", s)
	fn()
	t, _ := context.WithTimeout(context.Background(), time.Second*3)
	Periodically(t, time.Second*1, fn)
	// Output:
}
