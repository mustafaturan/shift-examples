package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mustafaturan/shift"
)

func main() {
	cb, err := shift.NewCircuitBreaker(
		"timeout-test",
		shift.WithInvocationTimeout(10*time.Millisecond),
	)
	if err != nil {
		panic(err)
	}

	var download shift.Operate = func(ctx context.Context) (interface{}, error) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://httpbin.org/range/2048?duration=8&chunk_size=128", nil)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		res, err := client.Do(req)
		log.Printf("res: %+v, err: %s", res, err)
		return res, err
	}

	ctx := context.Background()
	res, err := cb.Run(ctx, download)
	if err != nil {
		log.Printf("err: %s\n", err)
	}
	log.Printf("res: %+v", res)
	time.Sleep(1 * time.Second)
}
