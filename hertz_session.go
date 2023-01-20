package fastsession

import (
	"context"
	"log"
	"time"

	"github.com/josephGuo/fastsession/providers/memcache"
	"github.com/josephGuo/fastsession/providers/memory"
	"github.com/josephGuo/fastsession/providers/mysql"
	"github.com/josephGuo/fastsession/providers/postgre"
	"github.com/josephGuo/fastsession/providers/redis"
	"github.com/josephGuo/fastsession/providers/sqlite3"

	"github.com/cloudwego/hertz/pkg/app"
)

const DefaultKey = "github.com/josephGuo/fastsession"

func buildProvider(providerName string, cfg *Config) Provider {
	var provider Provider
	var err error

	switch providerName {
	case "memory":
		cfg.EncodeFunc = MSGPEncode
		cfg.DecodeFunc = MSGPDecode
		provider, err = memory.New(memory.Config{})
	case "redis":
		cfg.EncodeFunc = MSGPEncode
		cfg.DecodeFunc = MSGPDecode
		provider, err = redis.New(redis.Config{
			KeyPrefix:   "session",
			Addr:        "127.0.0.1:6379",
			PoolSize:    8,
			IdleTimeout: 30 * time.Second,
		})
	case "memcache":
		cfg.EncodeFunc = MSGPEncode
		cfg.DecodeFunc = MSGPDecode
		provider, err = memcache.New(memcache.Config{
			KeyPrefix: "session",
			ServerList: []string{
				"0.0.0.0:11211",
			},
			MaxIdleConns: 8,
		})
	case "mysql":
		cfg := mysql.NewConfigWith("127.0.0.1", 3306, "root", "session", "test", "session")
		provider, err = mysql.New(cfg)
	case "postgre":
		cfg := postgre.NewConfigWith("127.0.0.1", 5432, "postgres", "session", "test", "session")
		provider, err = postgre.New(cfg)
	case "sqlite3":
		cfg := sqlite3.NewConfigWith("test.db", "session")
		provider, err = sqlite3.New(cfg)
	default:
		panic("Invalid provider")
	}

	if err != nil {
		log.Fatal(err)
	}

	return provider
}

func NewHertzSession(providerName string, cfg Config) app.HandlerFunc {
	provider := buildProvider(providerName, &cfg)
	return func(ctx context.Context, c *app.RequestContext) {
		_, exist := c.Get(DefaultKey)
		if !exist {
			sessionManager := New(cfg)
			sessionManager.SetProvider(provider)
			c.Set(DefaultKey, sessionManager)
			log.Print("Starting example with provider: " + providerName)
		}
		log.Printf("before c.next handler index:%d handlers length:%d\n", c.GetIndex(), len(c.Handlers()))
		c.Next(ctx)
		log.Printf("after c.next handler index:%d handlers length:%d\n", c.GetIndex(), len(c.Handlers()))
	}
}

func DefaultSession(ctx *app.RequestContext) *Session {
	return ctx.MustGet(DefaultKey).(*Session)
}

func DefaultStore(ctx *app.RequestContext) *Store {
	session := ctx.MustGet(DefaultKey).(*Session)
	store, _ := session.Get(ctx)
	return store
}
