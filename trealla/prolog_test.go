package trealla

import (
	"context"
	"io"
	"testing"
)

func TestClose(t *testing.T) {
	pl, err := New()
	if err != nil {
		t.Fatal(err)
	}
	pl.Close()
	_, err = pl.QueryOnce(context.Background(), "true")
	if err != io.EOF {
		t.Error("unexpected error", err)
	}
}

func TestLeakCheck(t *testing.T) {
	check := func(goal string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			pl, err := New()
			if err != nil {
				t.Fatal(err)
			}
			defer pl.Close()

			pl.Register(ctx, "interop_simple", 1, func(pl Prolog, subquery Subquery, goal Term) Term {
				return Atom("interop_simple").Of(int64(42))
			})

			size := 0
			for i := 0; i < 2048; i++ {
				q := pl.Query(ctx, goal)
				for q.Next(ctx) {
				}
				if err := q.Err(); err != nil {
					t.Fatal(err, "iter=", i)
				}
				q.Close()

				current := pl.Stats().MemorySize
				if size == 0 {
					size = current
				}
				if current > size {
					t.Fatal("possible leak: memory grew to:", current, "initial:", size)
				}
			}
			t.Logf("goal: %s size: %d", goal, size)
		}
	}
	t.Run("true", check("true."))
	t.Run("between(1,3,X)", check("between(1,3,X)."))
	t.Run("output", check("write(stdout, abc), write(stderr, def) ; write(stdout, xyz), write(stderr, qux) ; 1=2."))

	t.Run("simple interop", check("interop_simple(X)"))
	// t.Run("complex interop", check("interop_query(X)"))
}
