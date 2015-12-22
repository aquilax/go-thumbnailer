package image

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/h2non/bimg"

	"github.com/pydima/go-thumbnailer/config"
)

var (
	MARKER_JPG = []byte{0xff, 0xd8}
	MARKER_PNG = []byte{0x89, 0x50}
	MARKER_GIF = []byte{0x47, 0x49}
)

type ImageType int

const (
	UNKNOWN ImageType = iota
	JPG
	PNG
	GIF
)

type Image struct {
	width  uint
	height uint
	path   string
}

type InvalidImage struct {
	err string
}

func (e InvalidImage) Error() string {
	return e.err
}

func constructName(prefix string, img []byte) string {
	switch ImageFormat(img) {
	case GIF:
		return prefix + ".gif"
	case JPG:
		return prefix + ".jpg"
	case PNG:
		return prefix + ".png"
	default:
		return ""
	}
}

func CheckExtension(n string) error {
	for _, ext := range config.Base.ValidExtensions {
		if strings.HasSuffix(strings.ToLower(n), ext) {
			return nil
		}
	}
	return InvalidImage{fmt.Sprintf("Extension %s is not supported.", filepath.Ext(n))}
}

func ImageFormat(img []byte) ImageType {
	if len(img) < 2 {
		return UNKNOWN
	}

	switch {
	case bytes.Equal(img[:2], MARKER_JPG):
		return JPG
	case bytes.Equal(img[:2], MARKER_PNG):
		return PNG
	case bytes.Equal(img[:2], MARKER_GIF):
		return GIF
	default:
		return UNKNOWN
	}
}

func ImageDimensions(img []byte) (width, height int, err error) {
	r := bytes.NewReader(img)
	conf, _, err := image.DecodeConfig(r)
	return conf.Width, conf.Height, err
}

// vips doesn't support gif natively, so have to convert it with slow standart library
func convertGifToPng(img []byte) ([]byte, error) {
	r := bytes.NewReader(img)
	i, err := gif.Decode(r)
	if err != nil {
		return nil, err
	}

	res := new(bytes.Buffer)
	err = png.Encode(res, i)
	if err != nil {
		return nil, err
	}

	return res.Bytes(), nil
}

func ProcessImage(img []byte, opts bimg.Options) (res []byte, err error) {
	img_t := ImageFormat(img)
	switch img_t {
	case UNKNOWN:
		return nil, fmt.Errorf("got unknown type")
	case GIF:
		img, err = convertGifToPng(img)
		if err != nil {
			return nil, err
		}
	}

	meta, err := bimg.Metadata(img)
	if err != nil {
		return nil, err
	}

	if !meta.Alpha {
		opts.Background = bimg.Color{R: 0, G: 0, B: 0}
	}

	return bimg.Resize(img, opts)
}

func CreateThumbnails(original []byte) (map[string][]byte, error) {
	opts := bimg.Options(config.Base.ImageParam)

	origThumbnail, err := ProcessImage(original, opts)
	if err != nil {
		return nil, err
	}
	imgs := make(map[string][]byte)
	imgs[constructName("original", origThumbnail)] = origThumbnail

	for k, v := range config.Base.Thumbnails {
		opts.Width = v[0]
		opts.Height = v[1]
		img, err := ProcessImage(origThumbnail, opts)
		if err != nil {
			return nil, err
		}
		imgs[constructName(k, img)] = img
	}

	return imgs, nil
}
