package mp4

import (
	"io"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

type rangeReader struct {
	ctx      context.Context
	req      *http.Request
	location *types.Location
	next     types.RequestHandler
}

func (rr *rangeReader) Range(start, length uint64) io.ReadCloser {
	newreq := copyRequest(rr.req)
	newreq.Header.Set("Range", utils.HTTPRange{Start: start, Length: length}.Range())
	var in, out = io.Pipe()
	flexible := utils.NewFlexibleResponseWriter(func(frw *utils.FlexibleResponseWriter) {
		if frw.Code != http.StatusPartialContent {
			_ = out.CloseWithError(errUnsatisfactoryResponse)
		}
		frw.BodyWriter = out
	})
	go func() {
		defer func() {
			if err := out.Close(); err != nil {
				rr.location.Logger.Errorf("handler.mp4[%p]: error on closing rangeReaders output: %s", rr.req, err)
			}
		}()
		rr.next.RequestHandle(rr.ctx, flexible, newreq, rr.location)
	}()

	return in
}

func (rr *rangeReader) RangeRead(start, length uint64) (io.ReadCloser, error) {
	return rr.Range(start, length), nil
}
