package db

import (
	"encoding/json"
	"fmt"
	pb "github.com/king8fisher/caddycfginjector/proto/caddycfginjector"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	a := assert.New(t)
	a.Equal(*caddyConf, CaddyConf{})
	a.Equal(true, IsConfEmpty(), "initially empty")

	t.Run("patchRoute", testPatchRoute)
	t.Run("patchRoute racing", testAddRouteRace)
}

func testPatchRoute(t *testing.T) {
	a := assert.New(t)
	a.Equal(*caddyConf, CaddyConf{})
	route0 := Route{
		Id:      "0",
		Handles: nil,
		Matches: nil,
	}
	patchRoute(route0)
	a.Equal(*caddyConf, CaddyConf{}, "adding to empty conf produces empty conf")
	err := resetConfToMinimumNonEmptyConf()
	a.Nil(err, "resetConfToMinimumNonEmptyConf() shouldn't result in error")
	a.NotEqual(*caddyConf, CaddyConf{}, "resetConfToMinimumNonEmptyConf() shouldn't result in an empty config")
	a.Equal(false, IsConfEmpty())
	patchRoute(route0)
	a.NotEqual(caddyConf, CaddyConf{}, "adding to non-empty conf produces non-empty conf")
	a.NotNil(caddyConf.Apps.Http.Servers.Myserver.Routes, "after adding a route can't be nil")
	a.Equal("0", (*caddyConf.Apps.Http.Servers.Myserver.Routes)[0].Id)
	patchRoute(route0) // Re-adding the same route
	a.Equal("0", (*caddyConf.Apps.Http.Servers.Myserver.Routes)[0].Id, "route '0' should still be at the same position")
	a.Equal(1, len(*caddyConf.Apps.Http.Servers.Myserver.Routes), "should contain just one record after re-adding the same route id")
	route1 := Route{
		Id:      "1",
		Handles: nil,
		Matches: nil,
	}
	patchRoute(route1)
	a.Equal(2, len(*caddyConf.Apps.Http.Servers.Myserver.Routes), "should contain 2 records after adding another route")
	patchRoute(route0)
	a.Equal(2, len(*caddyConf.Apps.Http.Servers.Myserver.Routes), "should contain 2 records after re-adding route '0'")
	a.Equal("0", (*caddyConf.Apps.Http.Servers.Myserver.Routes)[0].Id, "route '0' should still be at the same position")
	a.Equal("1", (*caddyConf.Apps.Http.Servers.Myserver.Routes)[1].Id, "route '1' should still be at the same position")
	//r, _ := json.MarshalIndent(caddyConf, "", "  ")
	//fmt.Println(string(r))
}

func testAddRouteRace(t *testing.T) {
	a := assert.New(t)
	a.Nil(resetConfToMinimumNonEmptyConf(), "should reset without error")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			AddRoute(&pb.Route{
				Id:      strconv.Itoa(idx),
				Handles: nil,
				Matches: nil,
			})
			wg.Done()
		}(i)
	}
	wg.Wait()
	a.Equal(10, len(*caddyConf.Apps.Http.Servers.Myserver.Routes), "should fill every distinct route id")
	r, _ := json.MarshalIndent(caddyConf, "", "  ")
	fmt.Println(string(r))
}
