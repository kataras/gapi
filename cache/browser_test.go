package cache_test

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/cache"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/httptest"
	"strconv"
	"testing"
	"time"
)

func TestNoCache(t *testing.T) {
	t.Parallel()
	app := iris.New()
	app.Get("/", cache.NoCache, func(ctx iris.Context) {
		ctx.WriteString("no_cache")
	})

	// tests
	e := httptest.New(t, app)

	r := e.GET("/").Expect().Status(httptest.StatusOK)
	r.Body().Equal("no_cache")
	r.Header(context.CacheControlHeaderKey).Equal(cache.CacheControlHeaderValue)
	r.Header(cache.PragmaHeaderKey).Equal(cache.PragmaNoCacheHeaderValue)
	r.Header(cache.ExpiresHeaderKey).Equal(cache.ExpiresNeverHeaderValue)
}

func TestStaticCache(t *testing.T) {
	t.Parallel()
	// test change the time format, which is not reccomended but can be done.
	app := iris.New().Configure(iris.WithTimeFormat("02 Jan 2006 15:04:05 GMT"))

	cacheDur := 30 * (24 * time.Hour)
	var expectedTime time.Time
	app.Get("/", cache.StaticCache(cacheDur), func(ctx iris.Context) {
		expectedTime = time.Now()
		ctx.WriteString("static_cache")
	})

	// tests
	e := httptest.New(t, app)
	r := e.GET("/").Expect().Status(httptest.StatusOK)
	r.Body().Equal("static_cache")

	r.Header(cache.ExpiresHeaderKey).Equal(expectedTime.Add(cacheDur).Format(app.ConfigurationReadOnly().GetTimeFormat()))
	cacheControlHeaderValue := "public, max-age=" + strconv.Itoa(int(cacheDur.Seconds()))
	r.Header(context.CacheControlHeaderKey).Equal(cacheControlHeaderValue)
}

func TestCache304(t *testing.T) {
	t.Parallel()
	app := iris.New()

	expiresEvery := 4 * time.Second
	app.Get("/", cache.Cache304(expiresEvery), func(ctx iris.Context) {
		ctx.WriteString("send")
	})
	// handlers
	e := httptest.New(t, app)

	// when 304, content type, content length and if ETagg is there are removed from the headers.
	insideCacheTimef := time.Now().Add(-expiresEvery).UTC().Format(app.ConfigurationReadOnly().GetTimeFormat())
	r := e.GET("/").WithHeader(context.IfModifiedSinceHeaderKey, insideCacheTimef).Expect().Status(httptest.StatusNotModified)
	r.Headers().NotContainsKey(context.ContentTypeHeaderKey).NotContainsKey(context.ContentLengthHeaderKey).NotContainsKey("ETag")
	r.Body().Equal("")

	// continue to the handler itself.
	cacheInvalidatedTimef := time.Now().Add(expiresEvery).UTC().Format(app.ConfigurationReadOnly().GetTimeFormat()) // after ~5seconds.
	r = e.GET("/").WithHeader(context.LastModifiedHeaderKey, cacheInvalidatedTimef).Expect().Status(httptest.StatusOK)
	r.Body().Equal("send")
	// now without header, it should continue to the handler itself as well.
	r = e.GET("/").Expect().Status(httptest.StatusOK)
	r.Body().Equal("send")
}
