package csvprocessor

import (
	"context"
)

type any = interface{} //nolint:predeclared
type csvCtx struct {
	context.Context //nolint:containedctx
	m               map[ctxKey]any
}

func newCtx() *csvCtx {
	ctx := csvCtx{}
	ctx.m = make(map[ctxKey]any)

	return &ctx
}

func (c *csvCtx) Value(key any) any {
	k, ok := key.(ctxKey)
	if !ok {
		return nil
	}

	return c.m[k]
}

func (c *csvCtx) setValue(key ctxKey, val any) {
	c.m[key] = val
}
