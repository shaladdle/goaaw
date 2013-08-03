// Package testutil provides some useful helpers for doing unit tests.
package testutil

import (
	crand "crypto/rand"
	mrand "math/rand"
    "io"
)

var (
	randReader io.Reader
)

type myRandReader struct {
	r *mrand.Rand
}

func (mrr *myRandReader) Read(b []byte) (int, error) {
	for i := 0; i < len(b); i++ {
		b[i] = byte(mrr.r.Intn(8))
	}

	return len(b), nil
}

func init() {
	randReader = crand.Reader
}

func SeedRNG(seed int64) {
	randReader = &myRandReader{mrand.New(mrand.NewSource(seed))}
}
