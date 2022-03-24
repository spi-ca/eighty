package serve

import (
	"io"
	"os"
)

type (
	// File defines a abstract file object,which provides caching hash.
	File interface {
		io.Closer
		io.Reader
		io.Seeker
		// Name returns abstract path of file.
		Name() string

		// Stat returns a os.FileInfo describing the named file.
		// If there is an error, it will be of type *os.PathError.
		Stat() (os.FileInfo, error)

		// Hash returns a file content checksum that is used for the e-tags.
		Hash() []byte
	}

	// FileProvider is an alias of the file resolver function.
	FileProvider = func(name string) (File, error)
)
