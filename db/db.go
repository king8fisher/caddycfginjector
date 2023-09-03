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

type CaddyConf struct {
	Apps struct {
		Http struct {
			Servers struct {
				Myserver struct {
					AutomaticHttps struct {
						Skip []string `json:"skip"`
					} `json:"automatic_https"`
					Listen []string `json:"listen"`
					Routes *[]Route `json:"routes"`
				} `json:"myserver"`
			} `json:"servers"`
		} `json:"http"`
	} `json:"apps"`
}

func resetConfToEmpty() {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()
	caddyConf = &CaddyConf{}
}

func resetConfToMinimumNonEmptyConf() error {
	caddyConfMutex.Lock()
	defer caddyConfMutex.Unlock()
	v := `
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
	return json.Unmarshal([]byte(v), &caddyConf)
}

func IsConfEmpty() bool {
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
	if IsConfEmpty() {
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

type Upstream struct {
	Dial string `json:"dial"`
}

type Transport struct {
	Protocol string `json:"protocol"`
}

type Handle struct {
	Handler   string     `json:"handler"`
	Transport Transport  `json:"transport"`
	Upstreams []Upstream `json:"upstreams"`
}

type Match struct {
	Hosts []string `json:"host"`
	Paths []string `json:"path"`
}

type Route struct {
	Id      string   `json:"@id"`
	Handles []Handle `json:"handle"`
	Matches []Match  `json:"match"`
}

func handlerToString(handler pb.Handle_Handler) string {
	switch handler {
	case pb.Handle_ReverseProxy:
		return "reverse_proxy"
	default:
		slog.Error("unknown handler", "handler", handler.String())
		os.Exit(1)
	}
	return ""
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

func AddRoute(r *pb.Route) {
	var handles []Handle
	for _, h := range r.Handles {
		var upstreams []Upstream
		for _, u := range h.Upstreams {
			upstreams = append(
				upstreams,
				Upstream{Dial: fmt.Sprintf("%v:%v", u.Dial.Host, u.Dial.Port)})
		}
		handles = append(handles, Handle{
			Handler: handlerToString(h.Handler),
			Transport: Transport{
				Protocol: transportProtocolToString(h.Transport.Protocol),
			},
			Upstreams: upstreams,
		})
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
}
