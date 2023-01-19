package fastsession

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func newCookie() *cookie {
	return new(cookie)
}

func (c *cookie) get(ctx *app.RequestContext, name string) []byte {
	return ctx.Request.Header.Cookie(name)
}

func (c *cookie) set(ctx *app.RequestContext, name string, value []byte, domain string, expiration time.Duration, secure bool, sameSite protocol.CookieSameSite) {
	cookie := protocol.AcquireCookie()

	cookie.SetKey(name)
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	cookie.SetDomain(domain)
	cookie.SetValueBytes(value)
	cookie.SetSameSite(sameSite)

	if expiration >= 0 {
		if expiration == 0 {
			cookie.SetExpire(protocol.CookieExpireUnlimited)
		} else {
			cookie.SetExpire(time.Now().Add(expiration))
		}
	}

	if secure {
		cookie.SetSecure(true)
	}

	//ctx.Request.Header.SetCookieBytesKV(cookie.Key(), cookie.Value())
	ctx.Request.Header.SetCookie(name, string(value))
	ctx.Response.Header.SetCookie(cookie)

	protocol.ReleaseCookie(cookie)
}

func (c *cookie) delete(ctx *app.RequestContext, name string) {
	ctx.Request.Header.DelCookie(name)
	ctx.Response.Header.DelCookie(name)

	cookie := protocol.AcquireCookie()
	cookie.SetKey(name)
	cookie.SetValue("")
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	//RFC says 1 second, but let's do it 1 minute to make sure is working...
	exp := time.Now().Add(-1 * time.Minute)
	cookie.SetExpire(exp)
	ctx.Response.Header.SetCookie(cookie)

	protocol.ReleaseCookie(cookie)
}
