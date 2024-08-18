package traefik_auth_with_exceptions

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	goauth "github.com/abbot/go-http-auth"
)

type BasicAuth struct {
	Users     []string `json:"users,omitempty"`
	UsersFile string   `json:"usersFile,omitempty"`
	Realm     string   `json:"realm,omitempty"`
}

type Exceptions struct {
	IpList             []string `json:"ipList,omitempty"`
	HostList           []string `json:"hostList,omitempty"`
	HostUpdateInterval string   `json:"hostUpdateInterval,omitempty"`
}

type Config struct {
	BasicAuth     BasicAuth  `json:"basicAuth,omitempty"`
	Exceptions    Exceptions `json:"exceptions,omitempty"`
	AuthExtraTime string     `json:"authExtraTime,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type AuthWithExceptions struct {
	next          http.Handler
	checker       *ExceptionChecker
	users         map[string]string
	auth          *goauth.BasicAuth
	authExtraTime time.Duration
	name          string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Println("Plugin initializing...")

	// create exception checker
	checker := NewExceptionChecker(config.Exceptions)

	// collect basic auth users
	users, err := getUsers(config.BasicAuth.UsersFile, config.BasicAuth.Users, basicUserParser)
	if err != nil {
		return nil, err
	}

	// get auth extra time
	authExtraTime, err := time.ParseDuration(config.AuthExtraTime)
	if err != nil {
		fmt.Printf("Error parsing authExtraTime: %v\n", err)
		authExtraTime = time.Duration(0)
	}

	p := &AuthWithExceptions{
		next:          next,
		checker:       checker,
		users:         users,
		authExtraTime: authExtraTime,
		name:          name,
	}

	realm := defaultRealm
	if len(config.BasicAuth.Realm) > 0 {
		realm = config.BasicAuth.Realm
	}

	p.auth = &goauth.BasicAuth{Realm: realm, Secrets: p.secretBasic}

	fmt.Println("Plugin initialized")

	return p, nil
}

func (p *AuthWithExceptions) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Checking remote address: %s\n", req.RemoteAddr)

	reason := "exception"
	trusted := p.checker.IsTrustedRemoteAddr(req.RemoteAddr)

	if !trusted {
		user, password, ok := req.BasicAuth()

		if ok {
			secret := p.auth.Secrets(user, p.auth.Realm)
			if secret == "" || !goauth.CheckSecret(password, secret) {
				ok = false
			}
		}

		if p.authExtraTime > 0 {
			time.Sleep(p.authExtraTime)
		}

		if !ok {
			fmt.Println("Authentication failed")
			p.auth.RequireAuth(rw, req)
			return
		}

		req.URL.User = url.User(user)

		reason = "basic auth"
	}

	fmt.Printf("Authentication successfully (%s)\n", reason)

	p.next.ServeHTTP(rw, req)
}

func (p *AuthWithExceptions) secretBasic(user, realm string) string {
	if secret, ok := p.users[user]; ok {
		return secret
	}

	return ""
}

func basicUserParser(user string) (string, string, error) {
	split := strings.Split(user, ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("error parsing BasicUser: %v", user)
	}
	return split[0], split[1], nil
}
