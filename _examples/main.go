package main

import (
	"flag"
	"log"
	"time"

	"github.com/josephGuo/fastsession/providers/memcache"
	"github.com/josephGuo/fastsession/providers/memory"
	"github.com/josephGuo/fastsession/providers/mysql"
	"github.com/josephGuo/fastsession/providers/postgre"
	"github.com/josephGuo/fastsession/providers/redis"
	"github.com/josephGuo/fastsession/providers/sqlite3"

	"github.com/josephGuo/fastsession"

	"github.com/cloudwego/hertz/pkg/app/server"
)

const defaultProvider = "memory"

var session *fastsession.Session

func init() {
	providerName := flag.String("provider", defaultProvider, "Name of provider")
	flag.Parse()

	encoder := fastsession.Base64Encode
	decoder := fastsession.Base64Decode

	var provider fastsession.Provider
	var err error

	switch *providerName {
	case "memory":
		encoder = fastsession.MSGPEncode
		decoder = fastsession.MSGPDecode
		provider, err = memory.New(memory.Config{})
	case "redis":
		encoder = fastsession.MSGPEncode
		decoder = fastsession.MSGPDecode
		provider, err = redis.New(redis.Config{
			KeyPrefix:   "session",
			Addr:        "127.0.0.1:6379",
			PoolSize:    8,
			IdleTimeout: 30 * time.Second,
		})
	case "memcache":
		encoder = fastsession.MSGPEncode
		decoder = fastsession.MSGPDecode
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

	cfg := fastsession.NewDefaultConfig()
	cfg.EncodeFunc = encoder
	cfg.DecodeFunc = decoder
	cfg.CookieName = "casdoor_session_id"
	session = fastsession.New(cfg)

	if err = session.SetProvider(provider); err != nil {
		log.Fatal(err)
	}

	log.Print("Starting example with provider: " + *providerName)
}

func main() {
	h := server.Default(server.WithHostPorts("127.0.0.1:8086"), server.WithExitWaitTime(3*time.Second))

	r := h.Engine
	r.GET("/", indexHandler)
	r.GET("/set", setHandler)
	r.GET("/get", getHandler)
	r.GET("/delete", deleteHandler)
	r.GET("/getAll", getAllHandler)
	r.GET("/flush", flushHandler)
	r.GET("/destroy", destroyHandler)
	r.GET("/sessionid", sessionIDHandler)
	r.GET("/regenerate", regenerateHandler)
	r.GET("/setexpiration", setExpirationHandler)
	r.GET("/getexpiration", getExpirationHandler)

	h.Spin()
}
