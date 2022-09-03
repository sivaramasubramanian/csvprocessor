package csvprocessor

import (
	"testing"
)

func Test_newCtx(t *testing.T) {
	t.Run("valid context test", func(t *testing.T) {
		got := newCtx()
		if got == nil {
			t.Errorf("newCtx() expected non-nil value, got = %v", got)
		}

		// check if all context methods are callable without null pointer error
		_, _ = got.Deadline()
		_ = got.Done()
		_ = got.Value(CtxChunkNum)
		_ = got.Err() //nolint:errcheck

		val := got.Value("CtxChunkNum")
		if val != nil {
			t.Errorf("csvCtx.Value() expected nil value, got = %v", got)
		}
	})
}
