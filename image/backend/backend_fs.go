package backend

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func imageGC(tmpDir string) error {
	dirs, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	for _, d := range dirs {
		name := d.Name()
		splitName := strings.Split(name, "_")
		if len(splitName) < 2 {
			continue
		}

		t, err := time.Parse(time.RFC3339, splitName[0])
		if err != nil {
			continue
		}

		if time.Since(t).Hours() > 24 {
			if err = os.RemoveAll(filepath.Join(tmpDir, name)); err != nil { // TODO: add logging.
				continue
			}
		}
	}

	return nil
}

type backendFS struct {
	BasePath string
	TmpDir   string
}

func (fb backendFS) generatePath(hash string) (path string) {
	return filepath.Join(fb.BasePath, hash[:2], hash[2:4], hash[4:6], hash[6:])
}

func (fb backendFS) exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (fb backendFS) generateDest(imgs map[string][]byte, basePath string) []string {
	res := make([]string, 0, len(imgs))
	for k := range imgs {
		res = append(res, filepath.Join(basePath, k))
	}
	return res
}

func (fb backendFS) createTmpDir() (tmpDir string, err error) {
	tmpDir, err = ioutil.TempDir(fb.TmpDir, time.Now().Format(time.RFC3339)+"_")

	if err != nil {
		pathErr, ok := err.(*os.PathError)
		if !ok {
			return
		}

		if pathErr.Err != syscall.ENOENT {
			return
		}

		if err = os.MkdirAll(fb.TmpDir, 0755); err != nil {
			return
		}

		tmpDir, err = ioutil.TempDir(fb.TmpDir, time.Now().Format(time.RFC3339)+"_")
		if err != nil {
			return
		}
	}
	return tmpDir, nil
}

func (fb backendFS) baseImage(imgs map[string][]byte) ([]byte, error) {
	var img []byte

	for k, v := range imgs {
		if strings.HasPrefix(k, "original") { // image might have different extensions e.g. 'jpg', 'png'
			img = v
			break
		}
	}

	if img == nil {
		return nil, fmt.Errorf("cannot find original image")
	}

	return img, nil
}

func (fb backendFS) moveFiles(src, dst string) (err error) {
	parentDir := filepath.Dir(dst)
	err = os.MkdirAll(parentDir, 0755)
	if err != nil {
		return
	}

	err = os.Rename(src, dst)
	if err != nil {
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.ENOTEMPTY {
			return &AlreadyExistsError{err: linkErr, Path: linkErr.New}
		}
	}
	return
}

func (fb backendFS) Save(imgs map[string][]byte) ([]string, error) {
	img, err := fb.baseImage(imgs)
	if err != nil {
		return nil, err
	}

	s := sha1.Sum(img)
	hash := base64.URLEncoding.EncodeToString(s[:])

	dir := fb.generatePath(hash)
	if fb.exists(dir) {
		return fb.generateDest(imgs, dir), nil
	}

	tmpDir, err := fb.createTmpDir()
	if err != nil {
		return nil, err
	}

	defer func() {
		if fb.exists(tmpDir) {
			os.RemoveAll(tmpDir)
		}
	}()

	for k, v := range imgs {
		if err := ioutil.WriteFile(filepath.Join(tmpDir, k), v, 0644); err != nil {
			return nil, err
		}
	}

	if err = fb.moveFiles(tmpDir, dir); err != nil {
		return nil, err
	}

	return fb.generateDest(imgs, dir), nil
}
