package common

import (
	"aaw/sync/util"
	"fmt"
	"io"
	"reflect"
)

const (
	putNetMsg = iota
	getNetMsg
	delNetMsg
	getIdxNetMsg
)

func readOneByte(r io.Reader) (byte, error) {
	ret := make([]byte, 1)
	n, err := r.Read(ret)
	if err != nil {
		return 0, err
	}

	if n != 1 {
		return 0, fmt.Errorf("did not read one byte")
	}

	return ret[0], nil
}

func writeOneByte(w io.Writer, b byte) error {
	n, err := w.Write([]byte{b})
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("did not write one byte")
	}

	return nil
}

func writeMsg(w io.Writer, msgType byte, msg interface{}) error {
	err := writeOneByte(w, msgType)
	if err != nil {
		return err
	}

	err = util.WriteObject(w, msg)
	if err != nil {
		return err
	}

	return nil
}

func ReadMsg(r io.Reader) (interface{}, error) {
	msgType, err := readOneByte(r)
	if err != nil {
		return nil, err
	}

	var msg interface{}

	switch msgType {
	case getNetMsg:
		msg = new(GetMsg)
	case putNetMsg:
		msg = new(PutMsg)
	case delNetMsg:
		msg = new(DelMsg)
	case getIdxNetMsg:
		msg = new(GetIdxMsg)
	default:
		return nil, fmt.Errorf("unrecognized message type")
	}

	err = util.ReadObject(r, msg)
	if err != nil {
		return nil, err
	}

	return reflect.ValueOf(msg).Elem().Interface(), nil
}

////////////////////////////////////////////////
// Get message
////////////////////////////////////////////////
type GetMsg struct {
	Id string
}

func WriteGetMsg(w io.Writer, id string) error {
	return writeMsg(w, getNetMsg, GetMsg{id})
}

////////////////////////////////////////////////
// Put message
////////////////////////////////////////////////
type PutMsg struct {
	Id string
}

func WritePutMsg(w io.Writer, id string) error {
	return writeMsg(w, putNetMsg, PutMsg{id})
}

////////////////////////////////////////////////
// Delete message
////////////////////////////////////////////////
type DelMsg struct {
	Id string
}

func WriteDelMsg(w io.Writer, id string) error {
	return writeMsg(w, delNetMsg, DelMsg{id})
}

////////////////////////////////////////////////
// Get Index message
////////////////////////////////////////////////
type GetIdxMsg struct {
}

func WriteGetIdxMsg(w io.Writer) error {
	return writeMsg(w, getIdxNetMsg, GetIdxMsg{})
}

func WriteIdx(w io.Writer, index map[string]FileInfo) error {
    err := util.WriteObject(w, index)
	if err != nil {
		return err
	}

	return nil
}

func ReadIdx(r io.Reader) (map[string]FileInfo, error) {
	var index map[string]FileInfo

	err := util.ReadObject(r, &index)
	if err != nil {
		return nil, err
	}

	return index, nil
}
