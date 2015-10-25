package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var Random *os.File

func init() {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatal(err)
	}
	Random = f
}

func UUID() string {
	b := make([]byte, 16)
	Random.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func DownloadImage(u string) (io.ReadCloser, error) {
	resp, err := http.Get(u)
	return resp.Body, err
}

func ReadImage(p string) (io.ReadCloser, error) {
	f, err := os.Open(p)
	return f, err
}
