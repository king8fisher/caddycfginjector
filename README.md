
gRPC server that accumulates configurations to update Caddy server.

Proof of concept with minimum API. 

The idea is quite simple:

* Caddy starts with a regular configuration but without routes:
  ```
  {
    "apps": {
      "http": {
        "servers": {
          "myserver": {
            "automatic_https": {
              "skip": []
            },
            "listen": [
              ":443"
            ],
            "routes": [
            ]
          }
        }
      }
    }
  }
  ``` 
  
  > Currently, recognition of emptiness is based on expecting "listen" and "skip" to be non-`null`. 

* This gRPC server is started.
* Each application uses this server by sending requests to register their routes via gRPC calls.
* This server then notifies Caddy of a change.

To ensure that the sequence in which all apps are started doesn't influence the outcome:

* This server does not accumulate any data until it gets a minimum configuration from Caddy.
  * Keeps querying Caddy until it responds. Never queries Caddy again about it.
    * Implication: If base configuration of Caddy is changed, this gRPC server will also have to be restarted.
* Each app that wants to register its route has to use the same `Route.id` (think domain name as a good candidate).
* Each app has to periodically announce itself with its route registration, since:
  * We can't request each app of its configuration.
  * Caddy can (re)start at any moment and should be able to receive each app's route from scratch. 
  * This gRPC server can also be restarted at any moment.

