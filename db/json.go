package db

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
