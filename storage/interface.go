// Package storage deals with files on the disk or whatever storage. It defines the
// Storage interface. It has methods for getting contents of a file, headers of a
// file and methods for removing files. Since every cache zone has its own storage
// it is possible to have different storage implementations running at the same
// time.
package storage

import (
	"io"
	"net/http"

	"github.com/ironsmile/nedomi/types"
)

// Storage represents a single unit of storage.
type Storage interface {
	// Returns a io.ReadCloser that will read from the `start`
	// of an object with ObjectId `id` to the `end`.
	Get(id types.ObjectID, start, end uint64) (io.ReadCloser, error)

	// Returns a io.ReadCloser that will read the whole file
	GetFullFile(id types.ObjectID) (io.ReadCloser, error)

	// Returns all headers for this object
	Headers(id types.ObjectID) (http.Header, error)

	// Discard an object from the storage
	Discard(id types.ObjectID) error

	// Discard an index of an Object from the storage
	DiscardIndex(index types.ObjectIndex) error
}
