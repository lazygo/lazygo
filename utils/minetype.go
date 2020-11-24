package utils

import (
	"mime"
	"path"
	"strings"
)

// 根据文件后缀获取文件的mimetype
func GetMimeType(name string) string {
	ext := strings.ToLower(path.Ext(name))
	return mime.TypeByExtension(ext)
}
