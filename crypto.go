package goutil

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
)

func Salt() (string, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return Sha1(token.String()), nil
}

func Md5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}

func Md5Bytes(bs []byte) []byte {
	h := md5.New()
	h.Write(bs)
	return h.Sum(nil)
}

func Md5StringToBytes(str string) []byte {
	h := md5.New()
	io.WriteString(h, str)
	return h.Sum(nil)
}

func Md5StringFromBytes(bytes []byte) string {
	src := md5.Sum(bytes)
	return strings.ToLower(hex.EncodeToString(src[:]))
}

func Md5FromReader(r io.Reader) string {
	h := md5.New()
	io.Copy(h, r)
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}

func Sha1(str string) string {
	h := sha1.New()
	io.WriteString(h, str)
	return hex.EncodeToString(h.Sum(nil))
}

func Sha1StringToBytes(str string) []byte {
	h := sha1.New()
	io.WriteString(h, str)
	return h.Sum(nil)
}

func Sha1Bytes(bs []byte) []byte {
	h := sha1.New()
	h.Write(bs)
	return h.Sum(nil)
}

func Sha1FromReader(r io.Reader) string {
	h := sha1.New()
	io.Copy(h, r)
	return hex.EncodeToString(h.Sum(nil))
}

// AESECBEncrypt encrypts string to base64 crypto using AES
func AESECBEncrypt(key []byte, text string) (string, error) {
	plaintext := []byte(text)

	//key = PKCS5Padding(key, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	plaintext = PKCS5Padding(plaintext, bs)
	if len(plaintext)%bs != 0 {
		return "", errors.New("EntryptAesECB Need a multiple of the blocksize")
	}

	dst := make([]byte, len(plaintext))
	mode := NewECBEncrypter(block)
	mode.CryptBlocks(dst, plaintext)

	// convert to base64
	return base64.StdEncoding.EncodeToString(dst), nil
}

// AESECBDecryptToBytes from base64 to decrypted bytes
func AESECBDecryptToBytes(key []byte, cryptoText string) ([]byte, error) {
	ciphertext, _ := base64.StdEncoding.DecodeString(cryptoText)

	//key = PKCS5Padding(key, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(ciphertext)%bs != 0 {
		return nil, errors.New("DecryptAesECB Need a multiple of the blocksize")
	}

	dst := make([]byte, len(ciphertext))

	mode := NewECBDecrypter(block)
	mode.CryptBlocks(dst, ciphertext)
	out := PKCS5UnPadding(dst)
	return out[:], nil
}

// AESECBDecrypt from base64 to decrypted string
func AESECBDecrypt(key []byte, cryptoText string) (string, error) {
	ciphertext, _ := base64.StdEncoding.DecodeString(cryptoText)

	//key = PKCS5Padding(key, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	if len(ciphertext)%bs != 0 {
		return "", errors.New("DecryptAesECB Need a multiple of the blocksize")
	}

	dst := make([]byte, len(ciphertext))

	mode := NewECBDecrypter(block)
	mode.CryptBlocks(dst, ciphertext)
	out := PKCS5UnPadding(dst)
	buf := bytes.NewBuffer(out[:])
	return buf.String(), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize //需要padding的数目
	//只要少于256就能放到一个byte中，默认的blockSize=16(即采用16*8=128, AES-128长的密钥)
	//最少填充1个byte，如果原文刚好是blocksize的整数倍，则再填充一个blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding) //生成填充的文本
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
