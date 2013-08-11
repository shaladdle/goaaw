package dedupfs

import (
	"encoding/json"
	"os"
	"strconv"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true
}

type blockMap map[uint64]block

func (bm blockMap) MarshalJSON() ([]byte, error) {
	newmap := make(map[string]block)
	for k, v := range bm {
		strkey := strconv.FormatUint(k, 10)
		newmap[strkey] = v
	}

	return json.Marshal(newmap)
}

func (bm blockMap) UnmarshalJSON(b []byte) error {
	newmap := make(map[string]block)
	err := json.Unmarshal(b, &newmap)
	if err != nil {
		return err
	}

	for k, v := range newmap {
		uintkey, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return err
		}
		bm[uintkey] = v
	}

	return nil
}
