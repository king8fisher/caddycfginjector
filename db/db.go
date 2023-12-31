package db

import (
	"encoding/json"
	"fmt"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"log/slog"
	"os"
	"slices"
	"sync"
)

var caddyConf = &CaddyConf{}
var caddyConfMutex sync.Mutex

// ReadCaddyConf sends string representation of
// a config unless empty.
func ReadCaddyConf() (string, error) {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()
	if isCaddyConfEmptyNonBlocking() {
		return "", fmt.Errorf("empty config")
	}
	b, err := json.Marshal(caddyConf)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func SetCaddyConf(conf []byte) error {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()
	var c CaddyConf
	err := json.Unmarshal(conf, &c)
	if err != nil {
		return fmt.Errorf("unable to fit conf: %v", err)
	} else {
		if isConfEmpty(c) {
			return fmt.Errorf("unable to set internal conf: seems empty")
		}
		caddyConf = &c
		return nil
	}
}

func resetConfToEmpty() {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()
	caddyConf = &CaddyConf{}
}

func resetConfToMinimumNonEmptyConf() {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()
	c := InitialCaddyConfig()
	caddyConf = &c
}

func InitialCaddyConfigSrc() string {
	return `
{
  "apps": {
    "http": {
      "servers": {
        "myserver": {
          "automatic_https": {
            "skip": []
          },
          "listen": [":443"],
          "routes": []
        }
      }
    }
  }
}
`
}

func InitialCaddyConfig() CaddyConf {
	var c CaddyConf
	v := InitialCaddyConfigSrc()
	// Asserting the absence of this error with tests
	_ = json.Unmarshal([]byte(v), &c)
	return c
}

func isConfEmpty(conf CaddyConf) bool {
	if conf.Apps.Http.Servers.Myserver.AutomaticHttps.Skip == nil {
		return true
	}
	if conf.Apps.Http.Servers.Myserver.Listen == nil {
		return true
	}
	return false
}

func isCaddyConfEmptyNonBlocking() bool {
	if caddyConf.Apps.Http.Servers.Myserver.AutomaticHttps.Skip == nil {
		return true
	}
	if caddyConf.Apps.Http.Servers.Myserver.Listen == nil {
		return true
	}
	return false
}

// patchRoute changes route element to the passed if its Route.Id matches existing Route.Id.
func patchRoute(r Route) {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()

	// Guard empty configuration
	if isCaddyConfEmptyNonBlocking() {
		return
	}

	var routes []Route

	if caddyConf.Apps.Http.Servers.Myserver.Routes == nil {
		routes = append(routes, r)
	} else {
		added := false
		for _, rr := range *caddyConf.Apps.Http.Servers.Myserver.Routes {
			if rr.Id == r.Id {
				routes = append(routes, r)
				added = true
			} else {
				routes = append(routes, rr)
			}
		}
		if !added {
			routes = append(routes, r)
		}
	}

	caddyConf.Apps.Http.Servers.Myserver.Routes = &routes
}

func transportProtocolToString(protocol pb.Transport_Protocol) string {
	switch protocol {
	case pb.Transport_HTTP:
		return "http"
	case pb.Transport_FastCGI:
		return "fastcgi"
	default:
		slog.Error("unknown transport protocol", "protocol", protocol.String())
		os.Exit(1)
	}
	return ""
}

func AddRoute(r *pb.Route) error {
	err := validateRoute(r)
	if err != nil {
		return err
	}
	var handles []Handle
	for _, h := range r.Handles {
		switch h := h.Handler.(type) {
		case *pb.Handle_ReverseProxy:
			var upstreams []Upstream
			for _, u := range h.ReverseProxy.Upstreams {
				upstreams = append(
					upstreams,
					Upstream{Dial: fmt.Sprintf("%v:%v", u.Dial.Host, u.Dial.Port)})
			}
			handles = append(handles, Handle{
				Handler: "reverse_proxy",
				Transport: Transport{
					Protocol: transportProtocolToString(h.ReverseProxy.Transport.Protocol),
				},
				Upstreams: upstreams,
			})
		default:
			slog.Warn("unknown handler type", "handler", h)
		}
	}
	var matches []Match
	for _, m := range r.Matches {
		matches = append(matches, Match{
			Hosts: slices.Clone(m.Hosts),
			Paths: slices.Clone(m.Paths),
		})
	}
	a := Route{
		Id:      r.Id,
		Handles: handles,
		Matches: matches,
	}
	patchRoute(a)
	return nil
}

func validateRoute(r *pb.Route) error {
	if r.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if len(r.Handles) == 0 {
		return fmt.Errorf("handles should contain at least one element")
	}
	return nil
}
