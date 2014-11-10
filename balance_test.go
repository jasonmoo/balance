package main

import (
	"crypto/rand"
	"crypto/sha512"
	"runtime"
	"testing"
)

func TestSHA(t *testing.T) {

	runtime.GOMAXPROCS(runtime.NumCPU())

	Balance(SHA)
}

func BenchmarkShaSum1kb(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = SHA()
	}

}

func SHA() error {
	h := sha512.New()
	var data [10 << 10]byte
	rand.Read(data[:])
	h.Write(data[:])
	h.Sum(nil)
	return nil
}
