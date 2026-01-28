package ginmw

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CompressLogger interface {
	Infow(msg string, keysAndValues ...interface{})
}

type Codec interface {
	Name() string
	WrapWriter(w io.Writer) (io.WriteCloser, error)
	WrapReader(r io.Reader) (io.ReadCloser, error)
}

type compressWriter struct {
	gin.ResponseWriter
	enc io.WriteCloser
}

func newCompressWriter(w gin.ResponseWriter, codec Codec) (*compressWriter, error) {
	enc, err := codec.WrapWriter(w)
	if err != nil {
		return nil, err
	}
	return &compressWriter{
		ResponseWriter: w,
		enc:            enc,
	}, nil
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.enc.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.Header().Del("Content-Length")
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.enc.Close()
}

type compressReader struct {
	io.ReadCloser
	dec io.ReadCloser
}

func newCompressReader(r io.ReadCloser, codec Codec) (*compressReader, error) {
	dec, err := codec.WrapReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{
		ReadCloser: r,
		dec:        dec,
	}, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.dec.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.dec.Close(); err != nil {
		return err
	}
	return c.ReadCloser.Close()
}

func GinContentEncoding(logger CompressLogger, codec Codec) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		acceptEncoding := ctx.Request.Header.Get("Accept-Encoding")
		supprotsGZIP := strings.Contains(acceptEncoding, codec.Name())

		if supprotsGZIP {
			logger.Infow("Начало кодирования", "codec", codec.Name())
			gw, err := newCompressWriter(ctx.Writer, codec)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to init encoder"})
				return
			}
			ctx.Header("Vary", "Accept-Encoding")
			ctx.Header("Content-Encoding", codec.Name())
			ctx.Writer = gw
			defer gw.Close()
		}

		encodingContent := ctx.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(encodingContent, codec.Name())
		if sendsGzip {
			logger.Infow("Начало декодирования", "codec", codec.Name())
			gr, err := newCompressReader(ctx.Request.Body, codec)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid encoded body"})
				return
			}
			ctx.Request.Body = gr
			ctx.Request.Header.Del("Content-Encoding")
			ctx.Request.ContentLength = -1
			defer gr.Close()
		}

		ctx.Next()
	}
}
