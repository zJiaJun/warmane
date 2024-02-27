package decode

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"io"
)

func ResponseBody(encoding string, respBody []byte) ([]byte, error) {
	switch encoding {
	case "br":
		return io.ReadAll(brotli.NewReader(bytes.NewBuffer(respBody)))
	case "gzip":
		gr, _ := gzip.NewReader(bytes.NewBuffer(respBody))
		return io.ReadAll(gr)
	case "deflate":
		zr := flate.NewReader(bytes.NewBuffer(respBody))
		defer func(zr io.ReadCloser) {
			_ = zr.Close()
		}(zr)
		return io.ReadAll(zr)
	default:
		return io.ReadAll(bytes.NewBuffer(respBody))
	}
}
