package util

import (
	"bytes"
	"reflect"
	"testing"
)

func TestStringReadWrite(t *testing.T) {
	b := &bytes.Buffer{}

	wantName := "/path/to/file.jpg"

	err := WriteString(b, wantName)
	if err != nil {
		t.Fatal("File name write error:", err)
	}

	actName, err := ReadString(b)
	if err != nil {
		t.Fatal("File name read error:", err)
	}

	if actName != wantName {
		t.Errorf("Name does not match. Expected %v, got %v", wantName, actName)
	}
}

func TestSizeReadWrite(t *testing.T) {
	b := &bytes.Buffer{}

	wantSize := int64(1024 * 1024 * 1024)

	err := WriteInt64(b, wantSize)
	if err != nil {
		t.Fatal("File size write error:", err)
	}

	actSize, err := ReadInt64(b)
	if err != nil {
		t.Fatal("File size read error:", err)
	}

	if actSize != wantSize {
		t.Errorf("Size does not match. Expected %v, got %v", wantSize, actSize)
	}
}

func TestObjectReadWrite(t *testing.T) {
	b := &bytes.Buffer{}

	wantMap := map[string]int{
		"adam": 26,
		"blue": 2,
		"red":  5,
	}

	err := WriteObject(b, wantMap)
	if err != nil {
		t.Fatal("File size write error:", err)
	}

	var actMap map[string]int
	err = ReadObject(b, &actMap)
	if err != nil {
		t.Fatal("File size read error:", err)
	}

	if !reflect.DeepEqual(wantMap, actMap) {
		t.Errorf("Map does not match. Expected %v, got %v", wantMap, actMap)
	}
}
