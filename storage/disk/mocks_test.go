package disk

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ironsmile/nedomi/types"
)

// Mock http handler

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for i := 0; i < 5; i++ {
		w.Header().Add(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("value-%d", i))
	}

	w.WriteHeader(200)
}

type fakeUpstream struct {
	types.Upstream
	responses map[string]fakeResponse
}

func (f *fakeUpstream) addFakeResponse(path string, fake fakeResponse) {
	if fake.Headers == nil {
		fake.Headers = make(http.Header)
	}
	if fake.Headers.Get("Content-Length") == "" {
		fake.Headers.Set("Content-Length", strconv.Itoa(len(fake.Response)))
	}
	f.responses[path] = fake
}

type fakeResponse struct {
	Status       string
	ResponseTime time.Duration
	Response     string
	Headers      http.Header
	err          error
}

func newFakeUpstream() *fakeUpstream {
	return &fakeUpstream{
		responses: make(map[string]fakeResponse),
	}
}

func (f *fakeUpstream) GetSize(path string) (int64, error) {
	fake, ok := f.responses[path]
	if !ok {
		return 0, nil // @todo fix
	}
	if fake.err != nil {
		return 0, fake.err
	}

	return int64(len(fake.Response)), nil
}

func (f *fakeUpstream) GetRequest(path string) (*http.Response, error) {
	end, _ := f.GetSize(path)
	return f.GetRequestPartial(path, 0, uint64(end))
}

func (f *fakeUpstream) GetRequestPartial(path string, start, end uint64) (*http.Response, error) {
	fake, ok := f.responses[path]
	if !ok {
		return nil, nil // @todo fix
	}
	if fake.ResponseTime > 0 {
		time.Sleep(fake.ResponseTime)
	}

	if fake.err != nil {
		return nil, fake.err
	}

	if length := uint64(len(fake.Response)); end > length {
		end = length
	}

	return &http.Response{
		Status: fake.Status,
		Body:   ioutil.NopCloser(bytes.NewBufferString(fake.Response[start:end])),
		Header: fake.Headers,
	}, nil
}

func (f *fakeUpstream) GetHeader(path string) (*http.Response, error) {
	fake, ok := f.responses[path]
	if !ok {
		return nil, nil // @todo fix
	}
	if fake.ResponseTime > 0 {
		time.Sleep(fake.ResponseTime)
	}

	if fake.err != nil {
		return nil, fake.err
	}
	return &http.Response{Header: fake.Headers}, nil

}

type countingUpstream struct {
	*fakeUpstream
	called int32
}

func (c *countingUpstream) GetRequestPartial(path string, start, end uint64) (*http.Response, error) {
	atomic.AddInt32(&c.called, 1)
	return c.fakeUpstream.GetRequestPartial(path, start, end)
}

type blockingUpstream struct {
	*fakeUpstream
	requestPartial chan chan struct{}
	requestHeader  chan chan struct{}
}

func newBlockingUpstream(upstream *fakeUpstream) *blockingUpstream {
	return &blockingUpstream{
		fakeUpstream:   upstream,
		requestPartial: make(chan chan struct{}),
		requestHeader:  make(chan chan struct{}),
	}
}

func (b *blockingUpstream) GetRequestPartial(path string, start, end uint64) (*http.Response, error) {
	ch := make(chan struct{})
	b.requestPartial <- ch
	<-ch
	return b.fakeUpstream.GetRequestPartial(path, start, end)
}

func (b *blockingUpstream) GetHeader(path string) (*http.Response, error) {
	ch := make(chan struct{})
	b.requestHeader <- ch
	<-ch
	return b.fakeUpstream.GetHeader(path)
}
