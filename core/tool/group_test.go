package tool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithContext(t *testing.T) {
	t.Run("nil context uses background", func(t *testing.T) {
		g := WithContext(nil)
		require.NotNil(t, g)
		ctx := g.Context()
		require.NotNil(t, ctx)
		require.NoError(t, ctx.Err())
	})

	t.Run("provided context is respected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		g := WithContext(ctx)
		derivedCtx := g.Context()
		require.NotNil(t, g)
		require.NotNil(t, derivedCtx)
		require.NoError(t, derivedCtx.Err())
		cancel()
		<-derivedCtx.Done()
		require.Error(t, derivedCtx.Err())
	})
}

func TestAnyGroup_WaitFirst(t *testing.T) {
	t.Run("empty group returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		err := g.WaitFirst()
		require.NoError(t, err)
	})

	t.Run("single success returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err := g.WaitFirst()
		require.NoError(t, err)
	})

	t.Run("single error returns that error", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return errors.New("only error")
		})
		err := g.WaitFirst()
		require.Error(t, err)
		require.Contains(t, err.Error(), "only error")
	})

	t.Run("first success cancels remaining goroutines", func(t *testing.T) {
		g := WithContext(context.Background())
		var cancelled atomic.Int32
		g.Go(func(ctx context.Context) error {
			time.Sleep(30 * time.Millisecond)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()
			cancelled.Add(1)
			return ctx.Err()
		})
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()
			cancelled.Add(1)
			return ctx.Err()
		})
		err := g.WaitFirst()
		require.NoError(t, err)
		require.Equal(t, int32(2), cancelled.Load())
	})

	t.Run("all fail returns joined errors", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return errors.New("error 1")
		})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 2")
		})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 3")
		})
		err := g.WaitFirst()
		require.Error(t, err)
		require.Contains(t, err.Error(), "error 1")
		require.Contains(t, err.Error(), "error 2")
		require.Contains(t, err.Error(), "error 3")
	})

	t.Run("partial failure with one success returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return errors.New("error 1")
		})
		g.Go(func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 3")
		})
		err := g.WaitFirst()
		require.NoError(t, err)
	})

	t.Run("parent context cancellation propagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g := WithContext(ctx)
		var started atomic.Int32
		g.Go(func(ctx context.Context) error {
			started.Add(1)
			<-ctx.Done()
			return ctx.Err()
		})
		g.Go(func(ctx context.Context) error {
			started.Add(1)
			<-ctx.Done()
			return ctx.Err()
		})
		time.Sleep(30 * time.Millisecond)
		cancel()
		err := g.WaitFirst()
		require.Error(t, err)
	})

	t.Run("first success cancels context for slow goroutines", func(t *testing.T) {
		g := WithContext(context.Background())
		var cancelled atomic.Int32
		g.Go(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				cancelled.Add(1)
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		})
		err := g.WaitFirst()
		require.NoError(t, err)
		require.Equal(t, int32(1), cancelled.Load())
	})

	t.Run("multiple concurrent successes do not deadlock", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		done := make(chan error, 1)
		go func() {
			done <- g.WaitFirst()
		}()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("WaitFirst deadlocked with multiple concurrent successes")
		}
	})
}

func TestAnyGroup_WaitAll(t *testing.T) {
	t.Run("empty group returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		err := g.WaitAll()
		require.NoError(t, err)
	})

	t.Run("all success returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return nil
		})
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err := g.WaitAll()
		require.NoError(t, err)
	})

	t.Run("any success returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return errors.New("error 1")
		})
		g.Go(func(ctx context.Context) error {
			return nil
		})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 3")
		})
		err := g.WaitAll()
		require.NoError(t, err)
	})

	t.Run("all fail returns joined errors", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return errors.New("error 1")
		})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 2")
		})
		err := g.WaitAll()
		require.Error(t, err)
		require.Contains(t, err.Error(), "error 1")
		require.Contains(t, err.Error(), "error 2")
	})

	t.Run("waits for all goroutines to complete", func(t *testing.T) {
		g := WithContext(context.Background())
		var counter atomic.Int32
		g.Go(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			counter.Add(1)
			return errors.New("error 1")
		})
		g.Go(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			counter.Add(1)
			return errors.New("error 2")
		})
		err := g.WaitAll()
		require.Error(t, err)
		require.Equal(t, int32(2), counter.Load())
	})

	t.Run("does not cancel on first success", func(t *testing.T) {
		g := WithContext(context.Background())
		var completed atomic.Int32
		g.Go(func(ctx context.Context) error {
			completed.Add(1)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			completed.Add(1)
			return nil
		})
		g.Go(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			completed.Add(1)
			return nil
		})
		err := g.WaitAll()
		require.NoError(t, err)
		require.Equal(t, int32(3), completed.Load())
	})

	t.Run("parent context cancellation propagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g := WithContext(ctx)
		var started atomic.Int32
		g.Go(func(ctx context.Context) error {
			started.Add(1)
			<-ctx.Done()
			return ctx.Err()
		})
		g.Go(func(ctx context.Context) error {
			started.Add(1)
			<-ctx.Done()
			return ctx.Err()
		})
		time.Sleep(30 * time.Millisecond)
		cancel()
		err := g.WaitAll()
		require.Error(t, err)
	})

	t.Run("single error returns that error", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return errors.New("only error")
		})
		err := g.WaitAll()
		require.Error(t, err)
		require.Contains(t, err.Error(), "only error")
	})

	t.Run("second call returns nil after WaitAll", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err := g.WaitAll()
		require.NoError(t, err)
		err = g.WaitAll()
		require.NoError(t, err)
	})
}

func TestAnyGroup_Reentrant(t *testing.T) {
	t.Run("second WaitFirst call returns nil", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err := g.WaitFirst()
		require.NoError(t, err)
		err = g.WaitFirst()
		require.NoError(t, err)
	})

	t.Run("can add new funcs after WaitFirst", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err := g.WaitFirst()
		require.NoError(t, err)
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err = g.WaitFirst()
		require.NoError(t, err)
	})

	t.Run("can add new funcs after WaitAll", func(t *testing.T) {
		g := WithContext(context.Background())
		g.Go(func(ctx context.Context) error {
			return nil
		})
		err := g.WaitAll()
		require.NoError(t, err)
		g.Go(func(ctx context.Context) error {
			return errors.New("error after reset")
		})
		err = g.WaitAll()
		require.Error(t, err)
		require.Contains(t, err.Error(), "error after reset")
	})
}
