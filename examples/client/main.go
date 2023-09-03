package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("did not connect", "err", err)
		os.Exit(1)
	}

	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	c := pb.NewCaddyCfgInjectorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.AddRoute(ctx, &pb.AddRouteRequest{Route: &pb.Route{
		Id: "example.com",
		Handles: []*pb.Handle{
			{
				Handler: pb.Handle_ReverseProxy,
				Transport: &pb.Transport{
					Protocol: pb.Transport_HTTP,
				},
				Upstreams: []*pb.Upstream{
					{
						Dial: &pb.Dial{
							Host: "127.0.0.1",
							Port: 8080,
						},
					},
				},
			},
		},
		Matches: []*pb.Match{
			{
				Hosts: []string{
					"example.com",
					"www.example.com",
					"beta.example.com",
				},
				Paths: []string{"/*"},
			},
		},
	},
	})
	if err != nil {
		slog.Error("could not add route", "err", err)
	}
	if r.GetResult() == pb.AddRouteReply_ok {
		slog.Info("answer", "message", r.GetMessage())
	} else if r.GetResult() == pb.AddRouteReply_error {
		slog.Error("answer", "message", r.GetMessage())
	}
}
