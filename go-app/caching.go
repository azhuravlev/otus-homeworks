package main

import (
	"fmt"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

const CacheLifeATime = 10 * time.Minute

var cacheStore *persistence.RedisStore

type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

type cachedWriter struct {
	gin.ResponseWriter
	status  int
	written bool
	store   persistence.CacheStore
	expire  time.Duration
	key     string
}

func newCachedWriter(store persistence.CacheStore, expire time.Duration, writer gin.ResponseWriter, key string) *cachedWriter {
	return &cachedWriter{writer, 0, false, store, expire, key}
}

func (w *cachedWriter) WriteHeader(code int) {
	w.status = code
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *cachedWriter) Status() int {
	return w.ResponseWriter.Status()
}

func (w *cachedWriter) Written() bool {
	return w.ResponseWriter.Written()
}

func (w *cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.ResponseWriter.Write(data)
	if err == nil {
		store := w.store
		var cache responseCache
		if err := store.Get(w.key, &cache); err == nil {
			data = append(cache.Data, data...)
		}

		//cache responses with a status code < 300
		if w.Status() < 300 {
			val := responseCache{
				w.Status(),
				w.Header(),
				data,
			}
			err = store.Set(w.key, val, w.expire)
			if err != nil {
				// need logger
			}
		}
	}
	return ret, err
}

func (w *cachedWriter) WriteString(data string) (n int, err error) {
	ret, err := w.ResponseWriter.WriteString(data)
	//cache responses with a status code < 300
	if err == nil && w.Status() < 300 {
		store := w.store
		val := responseCache{
			w.Status(),
			w.Header(),
			[]byte(data),
		}
		store.Set(w.key, val, w.expire)
	}
	return ret, err
}

func initCache() {
	if viper.GetBool("cache-enabled") {
		cacheStore = persistence.NewRedisCache(viper.GetString("redis"), "", CacheLifeATime)
	}
}

func cachePageIfValid(store persistence.CacheStore, expire time.Duration, valid func() bool, handle gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var respCache responseCache
		url := c.Request.URL
		key := cache.CreateKey(url.RequestURI())

		if !valid() {
			store.Delete(key)
		}

		if err := store.Get(key, &respCache); err != nil {
			if err != persistence.ErrCacheMiss {
				fmt.Errorf("Redis cache error: %s\n", err)
			}
			// replace writer
			writer := newCachedWriter(store, expire, c.Writer, key)
			c.Writer = writer
			handle(c)

			// Drop caches of aborted contexts
			if c.IsAborted() {
				store.Delete(key)
			}
		} else {
			c.Writer.WriteHeader(respCache.Status)
			for k, vals := range respCache.Header {
				for _, v := range vals {
					c.Writer.Header().Set(k, v)
				}
			}
			c.Writer.Write(respCache.Data)
		}
	}
}