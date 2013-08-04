package rpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"reflect"
	"testing"
)

var coders = []Coder{
	gobCoder{},
}

// TestCoders
func TestCoders(t *testing.T) {
	type tuple struct {
		A int
		B string
		C []int
	}

	// For types that will be transmitted from empty interface containers, we
	// have to register them with the gob package.
	gob.Register(tuple{})

	values := []interface{}{
		int(8),
		byte(120),
		rpcClass(33),
		float64(1234),
		"hi there",
		true,
		[]int{1, 4, 19, 12, 22},
		tuple{8, "heyhey", []int{9, 9, 1}},
		[]interface{}{tuple{8, "1", []int{4}}, 44, "hey what's up?", float64(23)},
	}

	for _, coder := range coders {
		buf := &bytes.Buffer{}
		tname := reflect.TypeOf(coder).Name()

		for _, want := range values {
			if err := coder.Encode(buf, want); err != nil {
				t.Errorf("test %v: encode error: %v", tname, err)
				continue
			}

			var gotPtr interface{} = reflect.New(reflect.TypeOf(want)).Interface()
			if err := coder.Decode(buf, gotPtr); err != nil {
				t.Errorf("test %v: decode error: %v", tname, err)
			}

			if got := reflect.ValueOf(gotPtr).Elem().Interface(); !reflect.DeepEqual(got, want) {
				t.Errorf("test %v: decoded value not right, got %v, want %v", tname, got, want)
			}
		}
	}
}

func TestCoderCliSrv(t *testing.T) {
	cc, sc := gobCoder{}, gobCoder{}

	r, w := io.Pipe()

	wanta := byte(tagRPC)
	wantb := rpcClass(rpcNorm)

	go func() {
		if err := cc.Encode(w, wanta); err != nil {
			t.Errorf("encode error: %v", err)
		}
		if err := cc.Encode(w, wantb); err != nil {
			t.Errorf("encode error: %v", err)
		}
	}()

	var (
		gota byte
		gotb rpcClass
	)
	if err := sc.Decode(r, &gota); err != nil {
		t.Errorf("decode error: %v", err)
	}
	if err := sc.Decode(r, &gotb); err != nil {
		t.Errorf("decode error: %v", err)
	}

	if wanta != gota {
		t.Errorf("got %v, want %v", gota, wanta)
	}
	if wantb != gotb {
		t.Errorf("got %v, want %v", gotb, wantb)
	}
}
