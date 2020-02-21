# go-cgi-session
Golang library for reading and writing sessions created with Perl CGI::Session

[![GoDoc](https://godoc.org/github.com/jeffdoubleyou/go-cgi-session?status.svg)](https://godoc.org/github.com/jeffdoubleyou/go-cgi-session)

## Motivation
I am working on a project that needs to maintain login sessions between a Perl CGI appliation and Beego API.  I could not find any existing modules to handle this, so I made this.

It is *NOT* well documented at this time.

## Features

### Storage Drivers

* Memcached

### Serializers

* Data::Dumper ( Default CGI )

### ID Generation

* MD5

## Beego usage

```
import(
	_ "github.com/jeffdoubleyou/go-cgi-session/beegoSessionProvider"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/session"
)

func main() {
	beego.BConfig.WebConfig.Session.SessionProvider = "cgi"
	beego.BConfig.WebConfig.Session.SessionProviderConfig = `{"driverConfig":{"servers":["10.64.98.74:11212"],"timeout":2}}`
	beego.BConfig.WebConfig.Session.SessionName = "CGISESSID"

	var sessionConfig = &session.ManagerConfig{
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
	}

	beego.GlobalSessions, _ = session.NewManager("cgi", sessionConfig)
	
	beego.Run()
}
```

I can't for the life of me get a custom provider to work without using beego.BConfig to set the params.  I think there must be some sort of bug there, where it initializes the custom, then the default session provider.

## Notes

There is a limtation to the storage at the moment.  No complex objects can be stored.

Since the project I'm working on uses only Memcached storage, this is all that I have built, but could easily add more.

