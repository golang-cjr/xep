package muc

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func UserMapping() (ret map[string]interface{}) {
	if f, err := os.Open(filepath.Join("static", "user-mapping.json")); err == nil {
		ret = make(map[string]interface{})
		err = json.NewDecoder(f).Decode(&ret)
	}
	return
}
