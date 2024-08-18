# Traefik Plugin: Auth With Exceptions

With this Traefik Middleware Plugin you can create a basic auth and define certain exceptions.
These can be IPs, IP ranges (CIDR) or hostnames. The host names are then resolved at definable intervals into IPs.

## Configuration

```yaml
http:
  middlewares:
    auth-with-exceptions:
      plugin:
        traefik-auth-with-exceptions:
          authExtraTime: 300ms
          basicAuth:
            users:
              - user:{SHA}3xFtZp37KYwrmWcRuHa2sKuEpmo=  # user:user
            usersFile: /etc/traefik/.htusers
          exceptions:
            ipList:
              - 127.0.0.1
              - 172.20.0.0/16
              - 192.168.240.0/24
            hostList:
              - example.com
            hostUpdateInterval: 5m
```

* **authExtraTime** : extra time to slow down auth if using md5 or sha hashed passwords. (e.g. 1s, 300ms)
* **basicAuth**:
  * **realm** : realm name
  * **users** : list of users (username:passwordhash)
  * **usersFile** : file with path to a users file
* **exceptions**:
  * **ipList** : List with IPs and CIDRs to exclude from basic auth
  * **hostList** : Hostnames to exclude from basic auth
  * **hostUpdateInterval** : Interval to update the IPs of hostList (e.g. 1h30m, 15m, 60s)
