package main

import (
	"context"
	"flag"
	"github.com/king8fisher/caddycfginjector/lib"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"time"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	fn := lib.Fn(*addr, &pb.Route{
		Id: "example.com",
		Handles: []*pb.Handle{
			{
				Handler: &pb.Handle_ReverseProxy{
					ReverseProxy: &pb.ReverseProxy{
						Transport: &pb.Transport{
							Protocol: pb.Transport_HTTP,
						},
						Upstreams: []*pb.Upstream{
							{
								Dial: &pb.Dial{
									Host: "localhost",
									Port: uint32(8080),
								},
							},
						},
					},
				},
			},
		},
		Matches: []*pb.Match{
			{
				Hosts: []string{"example.com", "beta.example.com"},
				Paths: []string{"/*"},
			},
		},
	})
	t, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	lib.Periodically(t, time.Second*2, fn)
}
