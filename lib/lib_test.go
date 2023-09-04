package lib

import (
	"fmt"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
)

func ExampleConfig() {
	s := &pb.Route{
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
	}
	fmt.Printf("%v\n", s.String())
	// Output:
	// id:"example.com" handles:{reverseProxy:{transport:{} upstreams:{dial:{host:"localhost" port:8080}}}} matches:{hosts:"example.com" hosts:"beta.example.com" paths:"/*"}
}
