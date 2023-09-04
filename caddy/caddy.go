package caddy

import (
	"context"
	"fmt"
	"github.com/king8fisher/caddycfginjector/db"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// PatchCaddyCh receives string representation of conf for caddy and
// will automatically send it to Caddy, provided PatchCaddy runs in the
// background.
var PatchCaddyCh = make(chan string, 1)

// PatchCaddy will run in background and wait for
// new conf sent to PatchCaddyCh.
func PatchCaddy(ctx context.Context, port int) {
	prev := ""
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-PatchCaddyCh:
			_, err := postCaddyConfig(port, c)
			if err != nil {
				slog.Error("patch caddy config", "err", err)
			} else {
				if prev != c {
					// Skip notifying for the same conf
					slog.Info("patch caddy success", "conf", c)
					prev = c
				}
			}
		}
	}
}

// PollCaddy performs polling Caddy Server for its conf and pushes an initial conf
// in case it returns empty conf and init is set to true.
func PollCaddy(ctx context.Context, port int, init bool) {
	t := time.NewTimer(time.Millisecond)
	errCnt := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			conf, err := readConfig(port)
			stop := false
			if err != nil {
				slog.Error("caddy response", "err", err)
			} else {
				if conf == "null" || conf == "null\n" {
					if init {
						slog.Info("attempting to inject initial config")
						_, err := postCaddyConfig(port, db.InitialCaddyConfigSrc())
						if err != nil {
							slog.Error("caddy initial conf response", "err", err)
						}
					}
					errCnt++
					if errCnt%10 == 1 {
						// Slowing down same error emission
						slog.Error("caddy initial conf empty, skipping incoming routes")
					}
				} else {
					err := db.SetCaddyConf([]byte(conf))
					if err != nil {
						slog.Error("caddy initial conf rejected", "conf", conf, "err", err)
					} else {
						slog.Info("caddy initial conf received", "conf", conf)
						stop = true
					}
				}
			}
			if stop {
				t.Stop()
			} else {
				t.Reset(time.Second * 2)
			}
		}
	}
}

func readConfig(port int) (string, error) {
	loadConfig, err := http.Get(fmt.Sprintf("http://localhost:%d/config", port))
	if err != nil {
		return "", err
	}
	defer loadConfig.Body.Close()
	b, err := io.ReadAll(loadConfig.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func postCaddyConfig(port int, cfg string) (string, error) {
	loadConfig, err := http.Post(fmt.Sprintf("http://localhost:%d/load", port), "application/json",
		strings.NewReader(cfg))
	if err != nil {
		return "", err
	}
	defer loadConfig.Body.Close()
	b, err := io.ReadAll(loadConfig.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
