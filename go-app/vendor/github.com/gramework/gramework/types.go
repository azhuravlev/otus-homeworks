// Copyright 2017-present Kirill Danshin and Gramework contributors
// Copyright 2019-present Highload LTD (UK CN: 11893420)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//

package gramework

import (
	"fmt"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"

	"github.com/apex/log"
	"github.com/gramework/utils/nocopy"
	"github.com/valyala/fasthttp"
)

type (
	ipList struct {
		list map[string]struct{}
		mu   *sync.RWMutex
	}

	suspect struct {
		hackAttempts int32
	}

	suspectsList struct {
		list map[string]*suspect
		mu   *sync.RWMutex
	}

	// App represents a gramework app
	App struct {
		defaultRouter             *Router
		domains                   map[string]*Router
		_                         [8]byte // callback
		firewall                  *firewall
		firewallInit              *sync.Once
		Flags                     *Flags
		flagsQueue                []Flag
		Logger                    log.Interface
		name                      string
		Settings                  Settings
		TLSEmails                 []string
		TLSPort                   uint16
		middlewares               []func(*Context)
		middlewaresAfterRequest   []func(*Context)
		preMiddlewares            []func(*Context)
		domainListLock            *sync.RWMutex
		middlewaresAfterRequestMu *sync.RWMutex
		middlewaresMu             *sync.RWMutex
		preMiddlewaresMu          *sync.RWMutex
		EnableFirewall            bool
		flagsRegistered           bool
		HandleUnknownDomains      bool
		seed                      uintptr
		cookieDomain              string
		cookiePath                string
		NoDefaultPanicHandler     bool
		PanicHandlerNoPoweredBy   bool
		PanicHandlerCustomLayout  string
		internalLog               *log.Entry

		cookieExpire time.Duration

		// Gramework Protection's max detections of suspect before ban
		maxHackAttempts *int32
		// Gramework Protection's protected endpoint prefixes
		protectedPrefixes map[string]struct{}
		// Gramework Protection's protected paths of endpoints
		protectedEndpoints map[string]struct{}
		// Gramework Protection's trusted ip list
		trustedIP *ipList
		// Gramework Protection's untrusted (banned) ip list
		untrustedIP *ipList
		// Gramework Protection's suspects ip list
		suspectedIP *suspectsList

		serverBase       *fasthttp.Server
		runningServers   []runningServerInfo
		runningServersMu *sync.Mutex

		behind Behind

		sanitizerPolicy *bluemonday.Policy

		DefaultCacheOptions *CacheOptions
	}

	// CacheOptions is a handler cache configuration structure.
	CacheOptions struct {
		// TTL is the time that cached response is valid
		TTL time.Duration
		// Cacheable function returns if current request is cacheable.
		// By deafult, any request with Authentication header or any Cookies will not be cached for security reasons.
		// If you want to cache responses for authorized users, please replace both Cacheable and CacheKey functions
		// to make sure that CacheKey includes something like session id.
		Cacheable func(ctx *Context) bool
		// CacheKey function returns the cache key for current request
		CacheKey func(ctx *Context) []byte

		// ReadCache allows for cache engine replacement. By default, gramework uses github.com/VictoriaMetrics/fastcache.
		// ReadCache returns the value and boolean if the value was found and still valid.
		ReadCache func(ctx *Context, key []byte) ([]byte, bool)
		// StoreCache allows for cache engine replacement. By default, gramework uses github.com/VictoriaMetrics/fastcache.
		StoreCache func(ctx *Context, key, value []byte, ttl time.Duration)

		// CacheableHeaders is a list of headers that gramework can cache.
		// Note, that if X-ABC is present both in cacheable and noncacheable header lists,
		// it will not be cached.
		CacheableHeaders []string // slice of canonical header names
		// NonCacheableHeaders is a list of headers that gramework can not cache.
		// Note, that if X-ABC is present both in cacheable and noncacheable header lists,
		// it will not be cached.
		NonCacheableHeaders []string
	}

	runningServerInfo struct {
		bind string
		srv  *fasthttp.Server
	}

	contextKey string

	// Context is a gramework request context
	Context struct {
		*fasthttp.RequestCtx
		nocopy    nocopy.NoCopy
		Logger    log.Interface
		App       *App
		auth      *Auth
		Cookies   Cookies
		requestID string

		middlewaresShouldStopProcessing bool
		subPrefixes                     []string
		middlewareKilledReq             bool
		writer                          func(p []byte) (int, error)
	}

	// GQLRequest is a GraphQL request structure
	GQLRequest struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	// Cookies handles a typical cookie storage
	Cookies struct {
		Storage map[string]string
		Mu      sync.RWMutex
	}

	// Settings for an App instance
	Settings struct {
		Firewall FirewallSettings
	}

	// FirewallSettings represents a new firewall settings.
	// Internal firewall representation copies this settings
	// atomically.
	FirewallSettings struct {
		// MaxReqPerMin is a max request per minute count
		MaxReqPerMin int64
		// BlockTimeout in seconds
		BlockTimeout int64
	}

	firewall struct {
		// Store a copy of current settings
		MaxReqPerMin        *int64
		BlockTimeout        *int64
		blockListMutex      sync.Mutex
		requestCounterMutex sync.Mutex
		blockList           map[string]int64
		requestCounter      map[string]int64
	}

	// Flags is a flags storage
	Flags struct {
		values map[string]Flag
	}

	// Flag is a flag representation
	Flag struct {
		Name        string
		Description string
		Default     string
		Value       *string
	}

	// Router handles internal handler conversion etc.
	Router struct {
		router      *router
		httprouter  *Router
		httpsrouter *Router
		root        *Router
		app         *App
		mu          sync.RWMutex

		rootHandler []staticHandler
	}

	// SubRouter handles subs registration
	// like app.Sub("v1").GET("someRoute", "hi")
	SubRouter struct {
		parent   routerable
		prefix   string
		prefixes []string
	}

	routerable interface {
		handleReg(method, route string, handler interface{}, prefixes []string)
		determineHandler(handler interface{}) func(*Context)
	}

	// RequestHandler describes a standard request handler type
	RequestHandler func(*Context)

	// RequestHandlerErr describes a standard request handler with error returned type
	RequestHandlerErr func(*Context) error

	// Auth is a struct that handles
	// context's basic auth features
	Auth struct {
		login string
		pass  string

		parsed bool
		// if error occurred during parsing,
		// it will be always returned for current
		// context
		err error

		ctx *Context
	}

	// HTML type used to determine prerendered strings
	// as an HTML and give proper content-type
	HTML string

	// JSON type used to determine prerendered strings
	// as an JSON and give proper content-type
	JSON string
)

// crazy hack to solve nocopy false positive
var _ = fmt.Sprintf("%v", func() interface{} {
	ctx := Context{}
	return &ctx.nocopy
}())
