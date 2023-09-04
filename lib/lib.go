package lib

import (
	"context"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

// Periodically is a helper that calls fn quickly and then repeats the call after
// refreshDelay duration until ctx requests a cancellation.
//
// Provided ctx is also passed down to the fn.
func Periodically(ctx context.Context, refreshDelay time.Duration, fn func(ctx context.Context)) {
	t := time.NewTimer(time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			fn(ctx)
			t.Reset(refreshDelay)
		}
	}
}

var logger = slog.Default()

// SetLogger replaces default slog.Default() logger for targeting logging from this package.
// slog.SetDefault() would be a global way of changing slog destination.
func SetLogger(l *slog.Logger) {
	logger = l
}

// Fn returns a function that attempts to add a route to caddycfginjector gRPC server via dialTarget.
//
// Most likely this function will have to be called at least once, and then repeatedly using Periodically so that Caddy will
// have a chance to register the route in case of a later start.
//
// Logging will be sent to a slog.Default() unless changed by SetLogger.
func Fn(dialTarget string, route *pb.Route) func(ctx context.Context) {
	var prevAddRouteReply int32 = -1
	fn := func(ctx context.Context) {
		conn, err := grpc.DialContext(ctx, dialTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Error("[caddycfginjector] did not connect", "err", err)
			return
		}

		defer func(conn *grpc.ClientConn) {
			_ = conn.Close()
		}(conn)
		c := pb.NewCaddyCfgInjectorClient(conn)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		r, err := c.AddRoute(ctx, &pb.AddRouteRequest{Route: route})
		if err != nil {
			slog.Error("[caddycfginjector] could not add route", "err", err)
			return
		}
		//if r.GetResult() == pb.AddRouteReply_ok {
		//} else if r.GetResult() == pb.AddRouteReply_error {
		if prevAddRouteReply != int32(r.GetResult()) {
			slog.Info("[caddycfginjector] reply", "message", r.GetMessage())
		}
		prevAddRouteReply = int32(r.GetResult())
	}
	return fn
}
