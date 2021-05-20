package utils

import (
	"crypto/md5"
	"encoding/hex"
)

//Md5sum md5sum
func Md5sum(data []byte) string {
	h := md5.New() //#nosec
	_, err := h.Write(data)
	if err != nil {
		return ""
	}
	d := hex.EncodeToString(h.Sum(nil))
	return d
}
