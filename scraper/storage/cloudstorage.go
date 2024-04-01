package storage

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gitub.com/zJiajun/warmane/constant"
	"hash"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CloudStorage struct {
	apiUrl   string
	uuid     string
	password string
}

func NewCloudStorage(apiUrl string, uuid string, password string) *CloudStorage {
	return &CloudStorage{apiUrl: apiUrl, uuid: uuid, password: password}
}

func (cs *CloudStorage) Init() error {

	return nil
}

func (cs *CloudStorage) Visited(requestID uint64) error {

	return nil
}

func (cs *CloudStorage) IsVisited(requestID uint64) (bool, error) {

	return false, nil
}

func (cs *CloudStorage) Cookies(u *url.URL) string {
	return cs.fetchCookies()
}

func (cs *CloudStorage) SetCookies(u *url.URL, cookies string) {

}

type cookieCloudBody struct {
	Uuid      string `json:"uuid,omitempty"`
	Encrypted string `json:"encrypted,omitempty"`
}

type cookieData struct {
	Data struct {
		Warmane []warmane `json:"warmane"`
	} `json:"cookie_data"`
	UpdateTime time.Time `json:"update_time"`
}

type warmane struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate"`
	HostOnly       bool    `json:"hostOnly"`
	HttpOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreId        string  `json:"storeId"`
	Value          string  `json:"value"`
}

func (cs *CloudStorage) fetchCookies() string {
	apiUrl := cs.apiUrl
	uuid := cs.uuid
	password := cs.password

	if apiUrl == "" || uuid == "" || password == "" {
		log.Fatalf("COOKIE_CLOUD_HOST, COOKIE_CLOUD_UUID and COOKIE_CLOUD_PASSWORD env must be set")
	}
	var data *cookieCloudBody
	res, err := http.Get(apiUrl + "/get/" + uuid)
	if err != nil {
		log.Fatalf("Failed to request server: %v", err)
	}
	if res.StatusCode != 200 {
		log.Fatalf("Server return status %d", res.StatusCode)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed to read server response: %v", err)
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalf("Failed to parse server response as json: %v", err)
	}
	keyPassword := md5String(uuid, "-", password)[:16]
	decrypted, err := decryptCryptoJsAesMsg(keyPassword, data.Encrypted)
	if err != nil {
		log.Fatalf("Failed to decrypt: %v", err)
	}
	fmt.Printf("Decrypted: %s\n", decrypted)
	var cookieData *cookieData
	err = json.Unmarshal(decrypted, &cookieData)
	if err != nil {
		log.Fatalf("Failed to parse cookie response as json: %v", err)
	}
	var b strings.Builder
	for _, v := range cookieData.Data.Warmane {
		for _, ck := range constant.CookieKeys {
			if v.Name != ck {
				continue
			}
			var expires time.Time
			if v.ExpirationDate > 0 {
				sec, dec := math.Modf(v.ExpirationDate)
				expires = time.Unix(int64(sec), int64(dec*(1e9)))
			}
			var maxAge int
			if v.Name == "_wM" {
				maxAge = 15552000
			} else if v.Name != "PHPSESSID" {
				maxAge = 604800
			}

			c := &http.Cookie{
				Name:     v.Name,
				Value:    v.Value,
				Path:     v.Path,
				Domain:   v.Domain,
				Expires:  expires,
				MaxAge:   maxAge,
				Secure:   v.Secure,
				HttpOnly: v.HttpOnly,
				SameSite: http.SameSiteDefaultMode,
			}
			b.WriteString(c.String())
		}
	}
	return b.String()
}

// Decrypt a CryptoJS.AES.encrypt(msg, password) encrypted msg.
// ciphertext is the result of CryptoJS.AES.encrypt(), which is the base64 string of
// "Salted__" + [8 bytes random salt] + [actual ciphertext].
// actual ciphertext is padded (make it's length align with block length) using Pkcs7.
// CryptoJS use a OpenSSL-compatible EVP_BytesToKey to derive (key,iv) from (password,salt),
// using md5 as hash type and 32 / 16 as length of key / block.
// See: https://stackoverflow.com/questions/35472396/how-does-cryptojs-get-an-iv-when-none-is-specified ,
// https://stackoverflow.com/questions/64797987/what-is-the-default-aes-config-in-crypto-js
func decryptCryptoJsAesMsg(password string, ciphertext string) ([]byte, error) {
	const keylen = 32
	const blocklen = 16
	rawEncrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode Encrypted: %v", err)
	}
	if len(rawEncrypted) < 17 || len(rawEncrypted)%blocklen != 0 || string(rawEncrypted[:8]) != "Salted__" {
		return nil, fmt.Errorf("invalid ciphertext")
	}
	salt := rawEncrypted[8:16]
	encrypted := rawEncrypted[16:]
	key, iv := bytesToKey(salt, []byte(password), md5.New(), keylen, blocklen)
	newCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %v", err)
	}
	cfbdec := cipher.NewCBCDecrypter(newCipher, iv)
	decrypted := make([]byte, len(encrypted))
	cfbdec.CryptBlocks(decrypted, encrypted)
	decrypted, err = pkcs7strip(decrypted, blocklen)
	if err != nil {
		return nil, fmt.Errorf("failed to strip pkcs7 paddings (password may be incorrect): %v", err)
	}
	return decrypted, nil
}

const (
	pkcs5SaltLen = 8
	aes256KeyLen = 32
)

// From https://github.com/walkert/go-evp .
// BytesToKey implements the Openssl EVP_BytesToKey logic.
// It takes the salt, data, a hash type and the key/block length used by that type.
// As such it differs considerably from the openssl method in C.
func bytesToKey(salt, data []byte, h hash.Hash, keyLen, blockLen int) (key, iv []byte) {
	saltLen := len(salt)
	if saltLen > 0 && saltLen != pkcs5SaltLen {
		panic(fmt.Sprintf("Salt length is %d, expected %d", saltLen, pkcs5SaltLen))
	}
	var (
		concat   []byte
		lastHash []byte
		totalLen = keyLen + blockLen
	)
	for ; len(concat) < totalLen; h.Reset() {
		// concatenate lastHash, data and salt and write them to the hash
		h.Write(append(lastHash, append(data, salt...)...))
		// passing nil to Sum() will return the current hash value
		lastHash = h.Sum(nil)
		// append lastHash to the running total bytes
		concat = append(concat, lastHash...)
	}
	return concat[:keyLen], concat[keyLen:totalLen]
}

// BytesToKeyAES256CBC implements the SHA256 version of EVP_BytesToKey using AES CBC
func bytesToKeyAES256CBC(salt, data []byte) (key []byte, iv []byte) {
	return bytesToKey(salt, data, sha256.New(), aes256KeyLen, aes.BlockSize)
}

// BytesToKeyAES256CBCMD5 implements the MD5 version of EVP_BytesToKey using AES CBC
func bytesToKeyAES256CBCMD5(salt, data []byte) (key []byte, iv []byte) {
	return bytesToKey(salt, data, md5.New(), aes256KeyLen, aes.BlockSize)
}

// return the MD5 hex hash string (lower-case) of input string(s)
func md5String(inputs ...string) string {
	keyHash := md5.New()
	for _, str := range inputs {
		io.WriteString(keyHash, str)
	}
	return hex.EncodeToString(keyHash.Sum(nil))
}

// from https://gist.github.com/nanmu42/b838acc10d393bc51cb861128ce7f89c .
// pkcs7strip remove pkcs7 padding
func pkcs7strip(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}
	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padLen], nil
}
