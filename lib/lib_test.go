package lib

import (
	"fmt"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"strings"
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
	// Somehow s.String() sometimes produces doubled spaces as separators. Can't rely on those
	fmt.Printf("%v\n", strings.Replace(s.String(), "  ", " ", -1))
	// Output:
	// id:"example.com" handles:{reverseProxy:{transport:{} upstreams:{dial:{host:"localhost" port:8080}}}} matches:{hosts:"example.com" hosts:"beta.example.com" paths:"/*"}
}
