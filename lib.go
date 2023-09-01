package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/king8fisher/caddycfginjector/internal"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	sl "github.com/srfrog/slices"
	"github.com/valyala/fasttemplate"
)

var jsonTemplate = `
{
  "apps": {
    "http": {
      "servers": {
        {{server_key}}: {
          "automatic_https": {
            "skip": []
          },
          "listen": [
            ":443"
          ],
          "routes": [
            {
              "@id": {{route_id}},
              "handle": [
                {
                  "handler": "reverse_proxy",
                  "transport": {
                    "protocol": "http"
                  },
                  "upstreams": [
                    {
                      "dial": {{host_port}}
                    }
                  ]
                }
              ],
              "match": [
                {
                  "host": [
                    {{matchHosts}}
                  ],
                  "path": [
                    {{matchPath}}
                  ]
                }
              ]
            }
          ]
        }
      }
    }
  }
}
`

// Config returns a json string with configuration for Caddy.
//
//   - serverKey - usually "myserver"
//   - routeId - unique ID within caddy configuration, ex. one of the domain names
//   - appHost / appPort - application host and port to proxy this route to
//   - matchHosts - list of matching hosts.
//   - matchPath - usually "/*"
func Config(serverKey string, routeId string, appHost string, appPort int, matchHosts []string, matchPath string) string {
	t := fasttemplate.New(jsonTemplate, "{{", "}}")
	s := t.ExecuteString(map[string]interface{}{
		"server_key": util.EncodeJSONString(serverKey),
		"route_id":   util.EncodeJSONString(routeId),
		"host_port":  util.EncodeJSONString(fmt.Sprintf("%v:%d", appHost, appPort)),
		"matchHosts": strings.Join(sl.Map(func(s string) string { return util.EncodeJSONString(s) }, matchHosts), ", "),
		"matchPath":  util.EncodeJSONString(matchPath),
	})
	return s
}

// Periodically is a helper that calls a function after refreshDelay duration and then repeats until ctx requests a cancellation.
func Periodically(ctx context.Context, refreshDelay time.Duration, fn func()) {
	ticker := time.NewTicker(refreshDelay)
	defer ticker.Stop()

	fn()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}

var logger = slog.Default()

// SetLogger replaces default slog.Default() logger for targeting logging from this package.
// slog.SetDefault() would be a global way of changing slog destination.
func SetLogger(l *slog.Logger) {
	logger = l
}

// Fn returns a function that executes configuration against caddy's loadURL (usually "http://localhost:2019/load").
//
// Most likely this function will have to be called at least once, and then repeatedly using Periodically so that Caddy will
// have a chance to pick up in case it (re)starts later.
//
// Logging will be sent to a slog.Default() unless changed by SetLogger.
func Fn(loadURL string, config string) func() {
	fn := func() {
		request, err := http.NewRequest("POST", loadURL, bytes.NewBufferString(config))
		if err != nil {
			logger.Error("creating caddy request", "err", err)
			return
		}
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			logger.Error("reading response", "err", err)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)

		body, _ := io.ReadAll(response.Body)
		if response.StatusCode != 200 {
			logger.Error("response !=200", "code", response.StatusCode, "body", string(body))
		}
	}
	return fn
}
