package disk

import (
	"io"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

const (
	headerFileName   = "headers"
	objectIDFileName = "objID"
)

// Disk implements the Storage interface by writing data to a disk
type Disk struct {
	partSize uint64 // actually uint32
	path     string
	logger   types.Logger
	/*
		cache          types.CacheAlgorithm
		indexRequests  chan *indexRequest
		headerRequests chan *headerRequest
		downloaded     chan *indexDownload
		removeChan     chan removeRequest
		closeCh        chan struct{}
		storageObjects uint64
	*/
}

// GetMetadata returns the metadata on disk for this object, if present.
func (s *Disk) GetMetadata(ctx context.Context, id types.ObjectID) (types.ObjectMetadata, error) {
	//!TODO: implement
	return types.ObjectMetadata{}, nil
}

// Get returns an io.ReadCloser that will read from the `start` of an object
// with ObjectId `id` to the `end`.
func (s *Disk) Get(ctx context.Context, id types.ObjectID, start, end uint64) (io.ReadCloser, error) {
	//!TODO: implement
	return nil, nil
}

// GetPart returns an io.ReadCloser that will read the specified part of the
// object from the disk.
func (s *Disk) GetPart(ctx context.Context, id types.ObjectIndex) (io.ReadCloser, error) {
	//!TODO: implement
	return nil, nil
}

// SaveMetadata writes the supplied metadata to the disk.
func (s *Disk) SaveMetadata(m types.ObjectMetadata) error {
	//!TODO: implement
	return nil
}

// SavePart writes the contents of the supplied object part to the disk.
func (s *Disk) SavePart(index types.ObjectIndex, data []byte) error {
	//!TODO: implement
	return nil
}

// Discard removes the object and its metadata from the disk.
func (s *Disk) Discard(id types.ObjectID) error {
	//!TODO: implement
	return nil
}

// DiscardPart removes the specified part of an Object from the disk.
func (s *Disk) DiscardPart(index types.ObjectIndex) error {
	//!TODO: implement
	return nil
}

// Walk iterates over the storage contents. It is used for restoring the
// state after the service is restarted.
func (s *Disk) Walk() <-chan types.ObjectFullMetadata {
	res := make(chan types.ObjectFullMetadata)
	//!TODO: implement
	close(res)
	return res
}

// New returns a new disk storage that ready for use.
func New(config *config.CacheZoneSection, log types.Logger) *Disk {
	storage := &Disk{
		partSize: config.PartSize.Bytes(),
		path:     config.Path,
		/*
			storageObjects: config.StorageObjects,
			cache:          ca,
			indexRequests:  make(chan *indexRequest),
			downloaded:     make(chan *indexDownload),
			removeChan:     make(chan removeRequest),
			headerRequests: make(chan *headerRequest),
			closeCh:        make(chan struct{}),
		*/
		logger: log,
	}
	/*
		go storage.loop()
		if err := storage.loadFromDisk(); err != nil {
			storage.logger.Error(err)
		}
		if err := storage.saveMetaToDisk(); err != nil {
			storage.logger.Error(err)
		}
	*/
	return storage
}

/*
type indexDownload struct {
	file        *os.File
	isCacheable bool
	index       types.ObjectIndex
	err         error
	requests    []*indexRequest
}

func (s *Disk) downloadIndex(ctx context.Context, index types.ObjectIndex) (*os.File, *http.Response, error) {
	vhost, ok := contexts.GetVhost(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("Could not get vhost from context.")
	}

	startOffset := uint64(index.Part) * s.partSize
	endOffset := startOffset + s.partSize - 1
	resp, err := vhost.Upstream.GetRequestPartial(index.ObjID.Path, startOffset, endOffset)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	file, err := CreateFile(path.Join(s.path, pathFromIndex(index)))
	if err != nil {
		return nil, nil, err
	}

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return nil, nil, utils.NewCompositeError(err, file.Close())
	}
	s.logger.Debugf("Storage [%p] downloaded for index %s with size %d", s, index, size)

	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, nil, utils.NewCompositeError(err, file.Close())
	}

	return file, resp, err
}

func (s *Disk) startDownloadIndex(request *indexRequest) *indexDownload {
	download := &indexDownload{
		index:    request.index,
		requests: []*indexRequest{request},
	}
	go func(ctx context.Context, download *indexDownload, index types.ObjectIndex) {
		file, resp, err := s.downloadIndex(ctx, index)
		if err != nil {
			download.err = err
		} else {
			download.file = file
			//!TODO: handle allowed cache duration
			download.isCacheable, _ = utils.IsResponseCacheable(resp)
			if download.isCacheable {
				s.writeObjectIDIfMissing(download.index.ObjID)
				//!TODO: don't do it for each piece and sanitize the headers
				s.writeHeaderToFile(download.index.ObjID, resp.Header)
			}
		}
		s.downloaded <- download
	}(request.context, download, request.index)
	return download
}

func (s *Disk) loop() {
	downloading := make(map[types.ObjectIndex]*indexDownload)
	headers := make(map[types.ObjectID]*headerQueue)
	headerFinished := make(chan *headerQueue)
	closing := false
	defer func() {
		close(headerFinished)
		close(s.downloaded)
		close(s.closeCh)
	}()
	for {
		select {
		case request := <-s.indexRequests:
			if request == nil {
				panic("request is nil")
			}
			s.logger.Debugf("Storage [%p]: downloading for indexRequest %+v\n", s, request)
			if download, ok := downloading[request.index]; ok {
				download.requests = append(download.requests, request)
				continue
			}
			if s.cache.Lookup(request.index) {
				file, err := os.Open(path.Join(s.path, pathFromIndex(request.index)))
				if err != nil {
					s.logger.Errorf("Error while opening file in cache: %s", err)
					downloading[request.index] = s.startDownloadIndex(request)
				} else {
					request.reader = file
					s.cache.PromoteObject(request.index)
					close(request.done)
				}
			} else {
				downloading[request.index] = s.startDownloadIndex(request)
			}

		case download := <-s.downloaded:
			delete(downloading, download.index)

			for _, request := range download.requests {
				if download.err != nil {
					s.logger.Errorf("Storage [%p]: error in downloading indexRequest %+v: %s\n", s, request, download.err)
					request.err = download.err
					close(request.done)
				} else {
					var err error
					request.reader, err = os.Open(download.file.Name()) //!TODO: optimize
					if err != nil {
						s.logger.Errorf("Storage [%p]: error on reopening just downloaded file for indexRequest %+v :%s\n", s, request, err)
						request.err = err
					}
					if download.isCacheable {
						s.cache.PromoteObject(request.index)
					}
					close(request.done)
				}
			}
			if !download.isCacheable || !s.cache.ShouldKeep(download.index) {
				syscall.Unlink(download.file.Name())
			}
			if closing && len(headers) == 0 && len(downloading) == 0 {
				return
			}

		case request := <-s.removeChan:
			s.logger.Debugf("Storage [%p] removing %s", s, request.path)
			request.err <- syscall.Unlink(request.path)
			close(request.err)

		// HEADERS
		case request := <-s.headerRequests:
			if queue, ok := headers[request.id]; ok {
				queue.requests = append(queue.requests, request)
				continue
			}
			header, err := s.readHeaderFromFile(request.id)
			if err == nil {
				request.header = header
				close(request.done)
				continue
			}
			//!TODO handle error

			queue := newHeaderQueue(request)
			headers[request.id] = queue
			go downloadHeaders(request.context, queue, headerFinished)

		case finished := <-headerFinished:
			delete(headers, finished.id)
			if finished.err == nil {
				if finished.isCacheable {
					//!TODO: do not save directly, run through the cache algo?
					s.writeObjectIDIfMissing(finished.id)
					s.writeHeaderToFile(finished.id, finished.header)
				}
			}
			for _, request := range finished.requests {
				if finished.err != nil {
					request.err = finished.err
				} else {
					request.header = finished.header // @todo copy ?
				}
				close(request.done)
			}

			if closing && len(headers) == 0 && len(downloading) == 0 {
				return
			}

		case <-s.closeCh:
			closing = true
			close(s.indexRequests)
			s.indexRequests = nil
			close(s.headerRequests)
			s.headerRequests = nil
			close(s.removeChan)
			s.removeChan = nil
			if len(headers) == 0 && len(downloading) == 0 {
				return
			}
		}
	}
}

//Writes the ObjectID to the disk in it's place if it already hasn't been written
func (s *Disk) writeObjectIDIfMissing(id types.ObjectID) error {
	pathToObjectID := path.Join(s.path, objectIDFileNameFromID(id))
	if err := os.MkdirAll(path.Dir(pathToObjectID), 0700); err != nil {
		s.logger.Errorf("Couldn't make directory for ObjectID [%s]: %s", id, err)
		return err
	}

	file, err := os.OpenFile(pathToObjectID, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(id)
}

func (s *Disk) readObjectIDForKeyNHash(key, hash string) (types.ObjectID, error) {
	file, err := os.Open(path.Join(s.path, key, hash, objectIDFileName))
	if err != nil {
		return types.ObjectID{}, err
	}

	var id types.ObjectID
	if err := json.NewDecoder(file).Decode(&id); err != nil {
		return types.ObjectID{}, err
	}

	return id, nil
}

type indexRequest struct {
	index   types.ObjectIndex
	reader  io.ReadCloser
	err     error
	done    chan struct{}
	context context.Context
}

func (ir *indexRequest) Close() error {
	<-ir.done
	if ir.err != nil {
		return ir.err
	}
	return ir.reader.Close()
}

func (ir *indexRequest) Read(p []byte) (int, error) {
	<-ir.done
	if ir.err != nil {
		return 0, ir.err
	}
	return ir.reader.Read(p)
}

// GetFullFile returns the whole file specified by the ObjectID
func (s *Disk) GetFullFile(ctx context.Context, id types.ObjectID) (io.ReadCloser, error) {
	vhost, ok := contexts.GetVhost(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not get vhost from context.")
	}

	size, err := vhost.Upstream.GetSize(id.Path)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		resp, err := vhost.Upstream.GetRequest(id.Path)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return s.Get(ctx, id, 0, uint64(size))
}

// Headers retunrs just the Headers for the specfied ObjectID
func (s *Disk) Headers(ctx context.Context, id types.ObjectID) (http.Header, error) {
	request := &headerRequest{
		id:      id,
		done:    make(chan struct{}),
		context: ctx,
	}
	s.headerRequests <- request
	<-request.done
	return request.header, request.err
}

// OldGet retuns an ObjectID from start to end
func (s *Disk) OldGet(ctx context.Context, id types.ObjectID, start, end uint64) (io.ReadCloser, error) {
	indexes := breakInIndexes(id, start, end, s.partSize)
	readers := make([]io.ReadCloser, len(indexes))
	for i, index := range indexes {
		request := &indexRequest{
			index:   index,
			done:    make(chan struct{}),
			context: ctx,
		}
		s.indexRequests <- request
		readers[i] = request
	}

	// work in start and end
	var startOffset, endLimit = start % s.partSize, end%s.partSize + 1
	readers[0] = newSkipReadCloser(readers[0], int(startOffset))
	readers[len(readers)-1] = newLimitReadCloser(readers[len(readers)-1], int(endLimit))

	return newMultiReadCloser(readers...), nil
}

func breakInIndexes(id types.ObjectID, start, end, partSize uint64) []types.ObjectIndex {
	firstIndex := start / partSize
	lastIndex := end/partSize + 1
	result := make([]types.ObjectIndex, 0, lastIndex-firstIndex)
	for i := firstIndex; i < lastIndex; i++ {
		result = append(result, types.ObjectIndex{
			ObjID: id,
			Part:  uint32(i),
		})
	}
	return result
}

type removeRequest struct {
	path string
	err  chan error
}

// OldDiscard a previosly cached ObjectID
func (s *Disk) OldDiscard(id types.ObjectID) error {
	request := removeRequest{
		path: path.Join(s.path, pathFromID(id)),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

// DiscardIndex a previosly cached ObjectIndex
func (s *Disk) DiscardIndex(index types.ObjectIndex) error {
	request := removeRequest{
		path: path.Join(s.path, pathFromIndex(index)),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

// GetCacheAlgorithm returns the used cache algorithm
func (s *Disk) GetCacheAlgorithm() *types.CacheAlgorithm {
	return &s.cache
}

// Close shuts down the Storage
func (s *Disk) Close() error {
	s.closeCh <- struct{}{}
	<-s.closeCh
	return nil
}
*/
