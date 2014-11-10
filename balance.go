package main

import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	Balance(SHA)
}

func SHA() error {
	h := sha512.New()
	var data [10 << 10]byte
	rand.Read(data[:])
	h.Write(data[:])
	h.Sum(nil)
	return nil
}

func Balance(f func() error) error {

	var (
		workers = int64(10)
		fail    = make(chan struct{})
		failed  = make(chan error)

		func_durations = make(chan time.Duration, 1024)
	)

	// worker count adjustment loop
	go func() {

		var (
			sample_duration = 500 * time.Millisecond
			tick            = time.NewTicker(sample_duration)

			executions,
			prev_executions,
			durations,
			duration_avg,
			prev_duration_avg int64
		)

		for {
			select {

			case d := <-func_durations:
				durations += d.Nanoseconds()
				executions++

			case <-tick.C:

				duration_avg = durations / executions

				fmt.Printf("workers: %d dur_avg: %d execs: %d ", atomic.LoadInt64(&workers), duration_avg, executions)

				// if we execute less and it takes longer per execution
				// decrease workers, otherwise increase
				if executions < prev_executions && duration_avg > prev_duration_avg {
					// inc := int64((float64(avg) / float64(prevavg)) * float64(workers))
					fmt.Println(" -")
					if n := atomic.AddInt64(&workers, -5); n == 0 {
						atomic.StoreInt64(&workers, 5)
					}
				} else {
					// proportional increase
					// inc := int64((float64(prevavg) / float64(avg)) * float64(workers))
					fmt.Println(" +")
					atomic.AddInt64(&workers, 5)
				}

				prev_duration_avg = duration_avg
				prev_executions, executions = executions, 0

			case <-fail:
				tick.Stop()
				return
			}
		}

	}()

	go func() {
		var (
			active int64
			rets   = make(chan error, 1024)
		)
		for {
			for active <= atomic.LoadInt64(&workers) {
				active++
				go func() {
					start := time.Now()
					rets <- f()
					func_durations <- time.Since(start)
				}()
			}
			select {
			case err := <-rets:
				if err != nil {
					fail <- struct{}{}
					failed <- err
					return
				}
				active--
			case <-time.After(10 * time.Millisecond):
			}
		}
	}()

	return <-failed

}
