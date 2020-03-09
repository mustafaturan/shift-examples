package main

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mustafaturan/shift"
	"github.com/mustafaturan/shift/restrictors"
)

func failureCounter(counter *int32) shift.OnFailure {
	return func(_ shift.State, _ error) {
		atomic.AddInt32(counter, 1)
	}
}

func failurePrinter() shift.OnFailure {
	return func(_ shift.State, err error) {
		log.Printf("Failed with %s", err)
	}
}

func successHandler(counter *int32) shift.OnSuccess {
	return func(_ interface{}) {
		atomic.AddInt32(counter, 1)
	}
}

func main() {
	restrictor, err := restrictors.NewConcurrentRunRestrictor("threshold-checker", 1)
	if err != nil {
		panic(err)
	}
	var failureCount, successCount int32
	cb, err := shift.NewCircuitBreaker(
		"timeout-test",
		shift.WithInvocationTimeout(2*time.Second),
		shift.WithRestrictors(restrictor),
		shift.WithOnFailureHandlers(failureCounter(&failureCount), failurePrinter()),
		shift.WithOnSuccessHandlers(successHandler(&successCount)),
	)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		go downloadAndPrint(cb, "http://httpbin.org/range/2048?duration=3&chunk_size=128")
	}

	time.Sleep(3 * time.Second)
	log.Printf("\n\nFailures: %d, Succeesses: %d", atomic.LoadInt32(&failureCount), atomic.LoadInt32(&successCount))
}

func download(url string) shift.Operate {
	return func(ctx context.Context) (interface{}, error) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		res, err := client.Do(req)
		log.Printf("response status code: %+v, err: %s", res.StatusCode, err)
		return res, err
	}
}

func downloadAndPrint(cb *shift.CircuitBreaker, url string) {
	ctx := context.Background()
	res, err := cb.Run(ctx, download(url))
	if err != nil {
		log.Printf("err: %s\n", err)
	}
	log.Printf("response: %+v", res)
}
