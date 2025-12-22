package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

type LocalCache struct {
	CacheOptions
}

func NewLocalCache(options CacheOptions) (*LocalCache, error) {
	if options.Path == "" {
		return nil, fmt.Errorf("CacheOptions must have Path set to an accessible location")
	}
	if err := os.MkdirAll(options.Path, 0o775); err != nil {
		return nil, fmt.Errorf("failed to create root cache directory %s: %s", options.Path, err)
	}

	return nil, &LocalCache{
		CacheOptions: options,
	}
}

func (cache *LocalCache) HasArtifacts(rule *Rule) bool {
	ruleDir := filepath.Join(cache.Path, rule.Checksum())

	return true
}

// TODO
//	StoreArtifacts(*Rule) error
//	ProcoduceArtifacts(*Rule) error
