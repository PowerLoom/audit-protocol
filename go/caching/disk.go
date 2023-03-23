package caching

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"
)

type LocalDiskCache struct{}

var _ DiskCache = (*LocalDiskCache)(nil)

func InitDiskCache() *LocalDiskCache {
	l := new(LocalDiskCache)
	err := gi.Inject(l)
	if err != nil {
		log.Fatal("Failed to inject disk cache", err)
	}

	return l
}

func (l LocalDiskCache) Read(filepath string) ([]byte, error) {
	// check if file exists
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		log.Errorf("file %s does not exist on disk", filepath)
		return nil, err
	}

	file, err := os.ReadFile(filepath)
	if err != nil {
		log.Errorf("error reading file %s from disk: %v", filepath, err)
		return nil, err
	}

	return file, nil
}

func (l LocalDiskCache) Write(filepath string, data []byte) error {
	err := os.WriteFile(filepath, data, 0644)
	if errors.Is(err, os.ErrNotExist) {
		// create the directory if it doesn't exist
		os.MkdirAll(filepath, 0700)
		err := os.WriteFile(filepath, data, 0644)
		if err != nil {
			log.Errorf("error writing file %s to disk: %v", filepath, err)

			return err
		}
	} else {
		log.Errorf("error writing file %s to disk: %v", filepath, err)

		return err
	}

	log.Debugf("Successfully wrote payload of size %d to file %s", len(data), filepath)

	return nil
}
