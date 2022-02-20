package helper

import (
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

// TODO implement as generic with go 1.18

func WriteHttpResponseToStream(c *gin.Context, res *http.Response) {
	r := io.Reader(res.Body)

	chanStream := make(chan string)
	go func() {
		defer close(chanStream)
		for {
			b := make([]byte, 2048)
			n, err := r.Read(b)
			if err != nil {
				chanStream <- "end"
				return
			} else {
				chanStream <- string(b[:n])
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		// Stream message to client from message channel
		if msg, ok := <-chanStream; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

func WriteImageBuildResponseToStream(c *gin.Context, res types.ImageBuildResponse) {
	r := io.Reader(res.Body)

	chanStream := make(chan string)
	go func() {
		defer close(chanStream)
		for {
			b := make([]byte, 2048)
			n, err := r.Read(b)
			if err != nil {
				chanStream <- "end"
				return
			} else {
				chanStream <- string(b[:n])
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		// Stream message to client from message channel
		if msg, ok := <-chanStream; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}
