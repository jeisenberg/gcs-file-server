package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	BucketName       = ""
	Protocol         = ""
	ProjectId        = ""
	GCSStorageId     = ""
	GCSStorageSecret = ""
	AES_KEY          = ""
)

type GCS struct {
	BucketName    string `json:"bucket_name"`
	Protocol      string `json:"protocol"`
	StorageId     string `json:"storage_id"`
	StorageSecret string `json:"storage_secret"`
	ProjectId     string `json:"project_id"`
}

type Config struct {
	GCS GCS `json:"GCS"`
}

type Response map[string]interface{}

func (r Response) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}

func init() {
	rtr := http.NewServeMux()
	rtr.HandleFunc("/media/", getMedia)
	rtr.HandleFunc("/_ah/health", ok)
	rtr.HandleFunc("/_ah/start", ok)

	http.Handle("/", rtr)
	initConfig()
}

func initConfig() error {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Println(err)
		return err
	}

	BucketName = config.GCS.BucketName
	Protocol = config.GCS.Protocol
	ProjectId = config.GCS.ProjectId
	GCSStorageId = config.GCS.StorageId
	GCSStorageSecret = config.GCS.StorageSecret
	return nil
}

// this is meant to signal a valid health check for container engine
func ok(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func getMedia(w http.ResponseWriter, r *http.Request) {

	fileName := strings.Split(r.URL.Path, "/media/")[1]
	canonicalResource, err := decodeFileName(fileName)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	// w.Write([]byte(fileName))
	// return

	timeHeader := time.Now()

	timeHeaderStr := timeHeader.Format(http.TimeFormat)

	message := strings.Join([]string{"GET", "", "application/xml", timeHeaderStr, ""}, "\n")
	message = strings.Join([]string{message, "/", BucketName, "/", canonicalResource}, "")

	key := []byte(GCSStorageSecret)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(message))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))

	encodedAsHeader := strings.Join([]string{"GOOG1 ", GCSStorageId, ":", encoded}, "")

	baseUrl := strings.Join([]string{Protocol, BucketName, ".storage.googleapis.com"}, "")

	getUrl := strings.Join([]string{baseUrl, canonicalResource}, "/")

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	c.Infof("%s", message)

	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	req.Header.Add("Authorization", encodedAsHeader)
	req.Header.Add("Date", timeHeaderStr)
	req.Header.Add("Content-Type", "application/xml")

	resp, err := client.Do(req)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	io.Copy(w, resp.Body)
}

func decodeFileName(encoded string) (string, error) {
	encoded, err := url.QueryUnescape(encoded)
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decryptFileName(encrypted string) (string, error) {
	block, err := aes.NewCipher([]byte(AES_KEY))
	if err != nil {
		return "", err
	}

	ciphertext := []byte("abcdef1234567890")
	iv := ciphertext[:aes.BlockSize]

	decrypter := cipher.NewCFBDecrypter(block, iv)

	decrypted := make([]byte, len(encrypted))

	decrypter.XORKeyStream(decrypted, []byte(encrypted))

	return base64.StdEncoding.EncodeToString(decrypted), nil
}
