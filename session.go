package fastsession

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

var (
	defaultLogger Logger = log.New(os.Stderr, "", log.LstdFlags)
)

// New returns a configured manager
func New(cfg Config) *Session {
	cfg.cookieLen = defaultCookieLen

	if cfg.CookieName == "" {
		cfg.CookieName = defaultSessionKeyName
	}

	if cfg.GCLifetime == 0 {
		cfg.GCLifetime = defaultGCLifetime
	}

	if cfg.SessionIDInURLQuery && cfg.SessionNameInURLQuery == "" {
		cfg.SessionNameInURLQuery = defaultSessionKeyName
	}

	if cfg.SessionIDInHTTPHeader && cfg.SessionNameInHTTPHeader == "" {
		cfg.SessionNameInHTTPHeader = defaultSessionKeyName
	}

	if cfg.SessionIDGeneratorFunc == nil {
		cfg.SessionIDGeneratorFunc = cfg.defaultSessionIDGenerator
	}

	if cfg.IsSecureFunc == nil {
		cfg.IsSecureFunc = cfg.defaultIsSecureFunc
	}

	if cfg.EncodeFunc == nil {
		cfg.EncodeFunc = Base64Encode
	}
	if cfg.DecodeFunc == nil {
		cfg.DecodeFunc = Base64Decode
	}

	if cfg.Logger == nil {
		cfg.Logger = defaultLogger
	}

	session := &Session{
		config: cfg,
		cookie: newCookie(),
		log:    cfg.Logger,
		storePool: &sync.Pool{
			New: func() interface{} {
				return NewStore()
			},
		},
		stopGCChan: make(chan struct{}),
	}

	return session
}

// SetProvider sets the session provider used by the sessions manager
func (s *Session) SetProvider(provider Provider) error {
	s.provider = provider

	if s.provider.NeedGC() {
		go s.startGC()
	}

	return nil
}

func (s *Session) startGC() {
	for {
		select {
		case <-time.After(s.config.GCLifetime):
			err := s.provider.GC()
			if err != nil {
				s.log.Printf("session GC crash: %v", err)
			}
		case <-s.stopGCChan:
			return
		}
	}
}

func (s *Session) stopGC() {
	s.stopGCChan <- struct{}{}
}

func (s *Session) setHTTPValues(ctx *app.RequestContext, sessionID []byte, expiration time.Duration) {
	secure := s.config.Secure && s.config.IsSecureFunc(ctx)
	s.cookie.set(ctx, s.config.CookieName, sessionID, s.config.Domain, expiration, secure, s.config.CookieSameSite)

	if s.config.SessionIDInHTTPHeader {
		//ctx.Request.Header.SetBytesV(s.config.SessionNameInHTTPHeader, sessionID)
		ctx.Request.Header.SetArgBytes([]byte(s.config.SessionNameInHTTPHeader), sessionID, true)
		ctx.Response.Header.SetBytesV(s.config.SessionNameInHTTPHeader, sessionID)
	}
}

func (s *Session) delHTTPValues(ctx *app.RequestContext) {
	s.cookie.delete(ctx, s.config.CookieName)

	if s.config.SessionIDInHTTPHeader {
		//ctx.Request.Header.Del(s.config.SessionNameInHTTPHeader)
		ctx.Request.Header.DelBytes([]byte(s.config.SessionNameInHTTPHeader))
		ctx.Response.Header.Del(s.config.SessionNameInHTTPHeader)
	}
}

// Returns the session id from:
// 1. cookie
// 2. http headers
// 3. query string
func (s *Session) getSessionID(ctx *app.RequestContext) []byte {
	val := ctx.Request.Header.Cookie(s.config.CookieName)
	if len(val) > 0 {
		return val
	}

	if s.config.SessionIDInHTTPHeader {
		val = ctx.Request.Header.Peek(s.config.SessionNameInHTTPHeader)
		if len(val) > 0 {
			return val
		}
	}

	if s.config.SessionIDInURLQuery {
		val = ctx.FormValue(s.config.SessionNameInURLQuery)
		if len(val) > 0 {
			return val
		}
	}

	return nil
}

// Get returns the user session
// if it does not exist, it will be generated
func (s *Session) Get(ctx *app.RequestContext) (*Store, error) {
	if s.provider == nil {
		return nil, errNotSetProvider
	}

	newUser := false

	id := s.getSessionID(ctx)
	if len(id) == 0 {
		id = s.config.SessionIDGeneratorFunc()
		if len(id) == 0 {
			return nil, errEmptySessionID
		}

		newUser = true
	}

	store := s.storePool.Get().(*Store)
	store.sessionID = id
	store.defaultExpiration = s.config.Expiration

	if !newUser {
		data, err := s.provider.Get(id)
		if err != nil {
			return nil, err
		}

		if err := s.config.DecodeFunc(&store.data, data); err != nil {
			return store, nil
		}
	}

	return store, nil
}

// Save saves the user session
//
// Warning: Don't use the store after exec this function, because, you will lose the after data
// For avoid it, defer this function in your request handler
func (s *Session) Save(ctx *app.RequestContext, store *Store) error {
	if s.provider == nil {
		return errNotSetProvider
	}

	id := store.GetSessionID()
	expiration := store.GetExpiration()

	providerExpiration := expiration
	if expiration == -1 {
		providerExpiration = keepAliveExpiration
	}

	data, err := s.config.EncodeFunc(store.GetAll())
	if err != nil {
		return err
	}

	if err := s.provider.Save(id, data, providerExpiration); err != nil {
		return err
	}

	s.setHTTPValues(ctx, id, expiration)

	store.Reset()
	s.storePool.Put(store)

	return nil
}

// Regenerate generates a new session id to the current user
func (s *Session) Regenerate(ctx *app.RequestContext) error {
	if s.provider == nil {
		return errNotSetProvider
	}

	id := s.getSessionID(ctx)
	expiration := s.config.Expiration

	newID := s.config.SessionIDGeneratorFunc()
	if len(newID) == 0 {
		return errEmptySessionID
	}

	providerExpiration := expiration
	if expiration == -1 {
		providerExpiration = keepAliveExpiration
	}

	if err := s.provider.Regenerate(id, newID, providerExpiration); err != nil {
		return err
	}

	s.setHTTPValues(ctx, newID, expiration)

	return nil
}

// Destroy destroys the session of the current user
func (s *Session) Destroy(ctx *app.RequestContext) error {
	if s.provider == nil {
		return errNotSetProvider
	}

	sessionID := s.getSessionID(ctx)
	if len(sessionID) == 0 {
		return nil
	}

	err := s.provider.Destroy(sessionID)
	if err != nil {
		return err
	}

	s.delHTTPValues(ctx)

	return nil
}
