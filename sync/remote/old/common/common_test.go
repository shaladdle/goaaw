package common

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestMessages(t *testing.T) {
	messages := []interface{}{
		PutMsg{"a"},
		GetMsg{"c"},
		DelMsg{"a"},
		PutMsg{"x"},
		DelMsg{"b"},
	}

	for _, msg := range messages {
		b := &bytes.Buffer{}

		switch msg := msg.(type) {
		case GetMsg:
			WriteGetMsg(b, msg.Id)
		case PutMsg:
			WritePutMsg(b, msg.Id)
		case DelMsg:
			WriteDelMsg(b, msg.Id)
		}

		rmsg, err := ReadMsg(b)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(msg, rmsg) {
			t.Fatalf("expected '%v', with type %v, got '%v' with type %v",
				msg, reflect.TypeOf(msg), rmsg, reflect.TypeOf(rmsg))
		}
	}
}

type dualPipe struct {
	r io.Reader
	w io.Writer
}

func (d *dualPipe) Read(p []byte) (int, error) {
	return d.r.Read(p)
}

func (d *dualPipe) Write(p []byte) (int, error) {
	return d.w.Write(p)
}

func makeDualPipe() (io.ReadWriter, io.ReadWriter) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	return &dualPipe{r1, w2}, &dualPipe{r2, w1}
}

func TestFileOps(t *testing.T) {
	c, s := makeDualPipe()

	wantStr := "this is the stuff"
	want := bytes.NewBufferString(wantStr)

	sendErr := make(chan error)
	go func() {
		sendErr <- sendFile(c, want, int64(want.Len()))
	}()

	act := &bytes.Buffer{}
	recvErr := make(chan error)
	go func() {
		recvErr <- recvFile(s, act)
	}()

	err := <-sendErr
	if err != nil {
		t.Fatal("sendFile failed:", err)
	}

	err = <-recvErr
	if err != nil {
		t.Fatal("recvFile failed:", err)
	}

	if act.String() != wantStr {
		t.Fatalf("strings do not match, expected '%v', got '%v'", wantStr, act.String())
	}
}
