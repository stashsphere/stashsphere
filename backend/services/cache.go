package services

import (
	"errors"
	"io"
	"os"
	"path"
)

type CacheService struct {
	cachePath string
}

func NewCacheService(cachePath string) (*CacheService, error) {
	err := os.MkdirAll(cachePath, 0755)
	if err != nil {
		return nil, err
	}
	return &CacheService{cachePath: cachePath}, nil
}

func (cs *CacheService) Put(key string, content []byte) error {
	filePath := path.Join(cs.cachePath, key)
	return os.WriteFile(filePath, content, 0600)
}

func (cs *CacheService) Get(key string) (io.Reader, error) {
	filePath := path.Join(cs.cachePath, key)
	return os.Open(filePath)
}

func (cs *CacheService) Delete(key string) error {
	filePath := path.Join(cs.cachePath, key)
	return os.Remove(filePath)
}

func (cs *CacheService) Exists(key string) bool {
	filePath := path.Join(cs.cachePath, key)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
