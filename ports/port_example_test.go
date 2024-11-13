package ports_test

import (
	"context"
	"fmt"

	"github.com/mcosta74/hexkit/ports"
)

func ExampleChain() {
	p := ports.Chain(
		annotate("1st"),
		annotate("2nd"),
		annotate("3rd"),
	)(myPort)

	if _, err := p(context.Background(), nil); err != nil {
		panic(err)
	}

	// Output:
	// 1st pre
	// 2nd pre
	// 3rd pre
	// hello!
	// 3rd post
	// 2nd post
	// 1st post
}

func annotate(s string) ports.Middleware[any, any] {
	return func(next ports.Port[any, any]) ports.Port[any, any] {
		return func(ctx context.Context, request any) (response any, err error) {
			fmt.Println(s, "pre")
			defer fmt.Println(s, "post")
			return next(ctx, request)
		}
	}
}

func myPort(context.Context, any) (any, error) {
	fmt.Println("hello!")
	return nil, nil
}
