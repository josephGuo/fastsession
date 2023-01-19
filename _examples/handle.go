package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// index handler
func indexHandler(_ context.Context, ctx *app.RequestContext) {
	html := "<h2>Welcome to use session, you should request to the: </h2>"

	html += `> <a href="/">/</a><br>`
	html += `> <a href="/set">set</a><br>`
	html += `> <a href="/get">get</a><br>`
	html += `> <a href="/delete">delete</a><br>`
	html += `> <a href="/getAll">getAll</a><br>`
	html += `> <a href="/flush">flush</a><br>`
	html += `> <a href="/destroy">destroy</a><br>`
	html += `> <a href="/sessionid">sessionid</a><br>`
	html += `> <a href="/regenerate">regenerate</a><br>`

	ctx.SetContentType("text/html;charset=utf-8")
	ctx.SetBodyString(html)
}

// set handler
func setHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
		foo := store.Get("foo")
		ctx.Response.AppendBodyString(fmt.Sprintf("\nafter store.Set-> session.Save then Get foo is nil:%t", foo == nil))
		store1, _ := session.Get(ctx)
		foo = store1.Get("foo")
		ctx.Response.AppendBodyString(fmt.Sprintf("\nre-run session.Get() obrain new store1  then Get foo is nil:%t", foo == nil))
	}()

	store.Set("foo", "bar")

	ctx.SetBodyString(fmt.Sprintf("Session SET: foo='%s' --> OK", store.Get("foo").(string)))
}

// get handler
func getHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
	}()

	val := store.Get("foo1")
	if val == nil {
		ctx.SetBodyString("Session GET: foo is nil")
		return
	}

	ctx.SetBodyString(fmt.Sprintf("Session GET: foo='%s'", val.(string)))
}

// delete handler
func deleteHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
	}()

	store.Delete("foo")

	val := store.Get("name")
	if val == nil {
		ctx.SetBodyString("Session DELETE: foo --> OK")
		return
	}
	ctx.SetBodyString("Session DELETE: foo --> ERROR")
}

// get all handler
func getAllHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
	}()

	store.Set("foo1", "bar1")
	store.Set("foo2", "2")
	store.Set("foo3", "bar3")
	store.Set("foo4", "bar4")

	data := store.GetAll()

	fmt.Println(data)
	var sb strings.Builder
	for k, v := range data.KV {
		sb.WriteString(k)
		sb.WriteByte(':')
		sb.WriteString(v.(string))
		sb.WriteByte('\n')
	}
	ctx.SetBodyString("Session GetAll: See the OS console!\n")
	ctx.Response.AppendBodyString(sb.String())
}

// flush handle
func flushHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
	}()

	store.Flush()

	data := store.GetAll()

	fmt.Println(data)

	ctx.SetBodyString("Session FLUSH: See the OS console!")
}

// destroy handle
func destroyHandler(_ context.Context, ctx *app.RequestContext) {
	err := session.Destroy(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}

	ctx.SetBodyString("Session DESTROY --> OK")
}

// get sessionID handle
func sessionIDHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
	}()

	sessionID := store.GetSessionID()
	ctx.SetBodyString("Session: Current session id: ")
	ctx.Write(sessionID)
}

// regenerate handler
func regenerateHandler(_ context.Context, ctx *app.RequestContext) {
	if err := session.Regenerate(ctx); err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}

	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}

	ctx.SetBodyString("Session REGENERATE: New session id: ")
	ctx.Write(store.GetSessionID())
}

// get expiration handler
func getExpirationHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}

	expiration := store.GetExpiration()

	ctx.SetBodyString("Session Expiration: ")
	ctx.WriteString(expiration.String())
}

// set expiration handler
func setExpirationHandler(_ context.Context, ctx *app.RequestContext) {
	store, err := session.Get(ctx)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := session.Save(ctx, store); err != nil {
			ctx.AbortWithError(consts.StatusInternalServerError, err)
		}
	}()

	err = store.SetExpiration(30 * time.Second)
	if err != nil {
		ctx.AbortWithError(consts.StatusInternalServerError, err)
		return
	}

	ctx.SetBodyString("Session Expiration set to 30 seconds")
}
