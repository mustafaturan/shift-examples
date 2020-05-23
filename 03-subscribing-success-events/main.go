package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mustafaturan/shift"
)

func failureHandler(name string) shift.OnFailure {
	return func(ctx context.Context, err error) {
		state := ctx.Value(shift.CtxState).(shift.State)
		stats := ctx.Value(shift.CtxStats).(shift.Stats)
		log.Printf("operation failed(%s) with error: %s, on state: %s with stats %+v", name, err, state, stats)
	}
}

func successHandler(name string) shift.OnSuccess {
	return func(ctx context.Context, res interface{}) {
		state := ctx.Value(shift.CtxState).(shift.State)
		stats := ctx.Value(shift.CtxStats).(shift.Stats)
		log.Printf("operation succeeded(%s) with error: %s, on state: %s with stats %+v", name, res, state, stats)
	}
}

func main() {
	cb, err := shift.New(
		"timeout-test",
		shift.WithInvocationTimeout(5*time.Second),
		shift.WithFailureHandlers(shift.StateClose, failureHandler("http-client")),
		shift.WithSuccessHandlers(shift.StateClose, successHandler("http-client")),
	)
	if err != nil {
		panic(err)
	}

	var download shift.Operate = func(ctx context.Context) (interface{}, error) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://httpbin.org/range/2048?duration=3&chunk_size=128", nil)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		res, err := client.Do(req)
		log.Printf("response status code: %+v, err: %s", res.StatusCode, err)
		return res, err
	}

	ctx := context.Background()
	res, err := cb.Run(ctx, download)
	if err != nil {
		log.Printf("err: %s\n", err)
	}
	log.Printf("response: %+v", res)
}
