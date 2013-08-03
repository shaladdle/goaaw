package util

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

var byteOrder = binary.LittleEndian

func WriteString(w io.Writer, s string) error {
	err := binary.Write(w, byteOrder, int64(len(s)))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}

func ReadString(r io.Reader) (string, error) {
	var slen int64
	err := binary.Read(r, byteOrder, &slen)
	if err != nil {
		return "", err
	}

	b := make([]byte, slen)
	_, err = r.Read(b)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func WriteInt64(w io.Writer, size int64) error {
	err := binary.Write(w, byteOrder, size)
	if err != nil {
		return err
	}

	return nil
}

func ReadInt64(r io.Reader) (int64, error) {
	var size int64
	err := binary.Read(r, byteOrder, &size)
	if err != nil {
		return -1, err
	}

	return size, nil
}

func WriteObject(w io.Writer, o interface{}) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	err = WriteInt64(w, int64(len(b)))
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func ReadObject(r io.Reader, o interface{}) error {
	size, err := ReadInt64(r)
	if err != nil {
		return err
	}

	b := make([]byte, size)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, o)
	if err != nil {
		return err
	}

	return nil
}
