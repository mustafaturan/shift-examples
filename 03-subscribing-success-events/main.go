package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mustafaturan/shift"
)

func failureHandler(name string) shift.OnFailure {
	return func(state shift.State, err error) {
		log.Printf("operation failed(%s) with error: %s, on state: %s", name, err, state)
	}
}

func successHandler(name string) shift.OnSuccess {
	return func(res interface{}) {
		log.Printf("operation succeeded for %s and response is %+v", name, res)
	}
}

func main() {
	cb, err := shift.NewCircuitBreaker(
		"timeout-test",
		shift.WithInvocationTimeout(5*time.Second),
		shift.WithOnFailureHandlers(failureHandler("http-client")),
		shift.WithOnSuccessHandlers(successHandler("http-client")),
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
