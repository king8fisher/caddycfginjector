package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/king8fisher/caddycfginjector/caddy"
	"github.com/king8fisher/caddycfginjector/db"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
	"os"
)

type server struct {
	pb.UnimplementedCaddyCfgInjectorServer
}

func (s *server) AddRoute(_ context.Context, in *pb.AddRouteRequest) (*pb.AddRouteReply, error) {
	err := db.AddRoute(in.Route)
	if err != nil {
		return &pb.AddRouteReply{
			Result:  pb.AddRouteReply_error,
			Message: err.Error(),
		}, nil
	}

	cf, err := db.ReadCaddyConf()
	if err == nil {
		caddy.PatchCaddyCh <- cf
		return &pb.AddRouteReply{
			Result:  pb.AddRouteReply_ok,
			Message: "ok",
		}, nil
	} else {
		return &pb.AddRouteReply{
			Result:  pb.AddRouteReply_error,
			Message: err.Error(),
		}, nil
		// return nil, status.Errorf(codes.Unimplemented, "method AddRoute not implemented")
	}
}

func main() {
	var host string
	flag.StringVar(&host, "host", "localhost", "Grpc server host. --host=\"\" to expose.")
	var port int
	flag.IntVar(&port, "port", 50051, "Grpc server port")
	var caddyPort int
	flag.IntVar(&caddyPort, "caddyPort", 2019, "Caddy port to poll and patch")
	var init bool
	flag.BoolVar(&init, "init", true, "Attempt to send initial conf to Caddy if returns empty")

	help := false
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(2)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}

	go caddy.PollCaddy(context.Background(), caddyPort, init)
	go caddy.PatchCaddy(context.Background(), caddyPort)

	s := grpc.NewServer()
	pb.RegisterCaddyCfgInjectorServer(s, &server{})
	slog.Info("caddycfginjector listens", "addr", lis.Addr())
	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve", "err", err)
		os.Exit(1)
	}
}
