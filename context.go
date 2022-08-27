package csvprocessor

import "time"

type any = interface{} //nolint:predeclared
type csvCtx struct {
	m map[ctxKey]any
}

func newCtx() *csvCtx {
	ctx := csvCtx{}
	ctx.m = make(map[ctxKey]any)

	return &ctx
}

func (c *csvCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *csvCtx) Done() <-chan struct{} {
	return nil
}

func (c *csvCtx) Err() error {
	return nil
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
