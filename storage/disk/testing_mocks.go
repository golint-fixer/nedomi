package disk

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

// Mock cache manager

type cacheManagerMock struct{}

func (c *cacheManagerMock) Lookup(o types.ObjectIndex) bool {
	return false
}

func (c *cacheManagerMock) ShouldKeep(o types.ObjectIndex) bool {
	return false
}

func (c *cacheManagerMock) AddObject(o types.ObjectIndex) error {
	return nil
}

func (c *cacheManagerMock) PromoteObject(o types.ObjectIndex) {}

func (c *cacheManagerMock) ConsumedSize() config.BytesSize {
	return 0
}

func (c *cacheManagerMock) ReplaceRemoveChannel(ch chan<- types.ObjectIndex) {

}

func (c *cacheManagerMock) Stats() types.CacheStats {
	return nil
}

// Mock http handler

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for i := 0; i < 5; i++ {
		w.Header().Add(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("value-%d", i))
	}

	w.WriteHeader(200)
}

type fakeUpstream struct {
	upstream.Upstream
	responses map[string]FakeResponse
}

func (f *fakeUpstream) addFakeResponse(path string, fake FakeResponse) {
	f.responses[path] = fake
}

type FakeResponse struct {
	Status       string
	ResponseTime time.Duration
	Response     string
	err          error
}

func NewFakeUpstream() *fakeUpstream {
	return &fakeUpstream{
		responses: make(map[string]FakeResponse),
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
	}, nil
}
