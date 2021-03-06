package mjpeg

import (
	"image"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"sync"
)

type Decoder struct {
	r *multipart.Reader
	m sync.Mutex
}

func NewDecoder(r io.Reader, b string) *Decoder {
	d := new(Decoder)
	d.r = multipart.NewReader(r, b)
	return d
}

func NewDecoderFromResponse(res *http.Response) (*Decoder, error) {
	_, param, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return NewDecoder(res.Body, param["boundary"]), nil
}

func NewDecoderFromURL(u string) (*Decoder, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return NewDecoderFromResponse(res)
}

func (d *Decoder) Decode(img *image.Image) error {
	p, err := d.r.NextPart()
	if err != nil {
		return err
	}
	tmp, err := jpeg.Decode(p)
	if err != nil {
		return err
	}
	d.m.Lock()
	*img = tmp
	d.m.Unlock()
	return nil
}
