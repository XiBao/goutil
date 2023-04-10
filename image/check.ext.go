package image

import (
	"mime/multipart"

	"github.com/h2non/filetype"
)

func CheckExtension(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "jpg", err
	}
	defer src.Close()
	head := make([]byte, 261)
	src.Read(head)
	kind, _ := filetype.Match(head)
	if kind == filetype.Unknown {
		return "jpg", nil
	}
	return kind.Extension, nil
}

func CheckExtensionWithBytes(data []byte) (string, error) {
	head := data[0:261]
	kind, _ := filetype.Match(head)
	if kind == filetype.Unknown {
		return "jpg", nil
	}
	return kind.Extension, nil
}
