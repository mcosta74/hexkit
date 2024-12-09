package requests_test

import (
	"context"
	"fmt"

	"github.com/mcosta74/hexkit/requests"
)

func ExampleChain() {
	p := requests.Chain(
		annotate("1st"),
		annotate("2nd"),
		annotate("3rd"),
	)(requests.HandlerFunc[any, any](handler))

	if _, err := p.Handle(context.Background(), nil); err != nil {
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

func annotate(s string) requests.Middleware[any, any] {
	return func(next requests.Handler[any, any]) requests.Handler[any, any] {
		return requests.HandlerFunc[any, any](func(ctx context.Context, request any) (response any, err error) {
			fmt.Println(s, "pre")
			defer fmt.Println(s, "post")
			return next.Handle(ctx, request)
		})
	}
}

func handler(context.Context, any) (any, error) {
	fmt.Println("hello!")
	return nil, nil
}
