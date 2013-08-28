package common

import (
	"github.com/shaladdle/goaaw/sync/util"
	"fmt"
	"io"
	"os"
)

type FileInfo struct {
	Id   string
	Size int64
}

func SendFile(conn io.ReadWriter, fname string) error {
	info, err := os.Stat(fname)
	if err != nil {
		return err
	}

	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	return sendFile(conn, f, info.Size())
}

func sendFile(conn io.ReadWriter, f io.Reader, size int64) error {
	err := util.WriteInt64(conn, size)
	if err != nil {
		return err
	}

	n, err := io.CopyN(conn, f, size)
	if err != nil {
		return err
	}

	if n != size {
		return fmt.Errorf("copy failed part way through, %v/%v", n, size)
	}

	nack, err := util.ReadInt64(conn)
	if err != nil {
		return err
	}

	if n != nack {
		return fmt.Errorf("ack was incorrect, expected %v, got %v", n, nack)
	}

	return nil
}

func RecvFile(conn io.ReadWriter, fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	return recvFile(conn, f)
}

func recvFile(conn io.ReadWriter, f io.Writer) error {
	n, err := util.ReadInt64(conn)
	if err != nil {
		return err
	}

	nact, err := io.CopyN(f, conn, n)
	if err != nil {
		return err
	}

	if n != nact {
		return fmt.Errorf("ack was incorrect, expected %v, got %v", n, nact)
	}

	err = util.WriteInt64(conn, nact)
	if err != nil {
		return err
	}

	return nil
}
