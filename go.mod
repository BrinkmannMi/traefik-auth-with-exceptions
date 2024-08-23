module github.com/BrinkmannMi/traefik-auth-with-exceptions

go 1.22.4

require github.com/abbot/go-http-auth v0.4.0

require golang.org/x/crypto v0.24.0 // indirect

replace github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
