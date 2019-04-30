package generate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sliver/server/assets"
	"sliver/server/db"
	"sliver/server/log"
	"strings"
	"time"
)

const (
	// sliverBucketName - Name of the bucket that stores data related to slivers
	sliverBucketName = "slivers"

	// sliverConfigNamespace - Namespace that contains sliver configs
	sliverConfigNamespace   = "config"
	sliverFileNamespace     = "file"
	sliverDatetimeNamespace = "datetime"
)

var (
	storageLog = log.NamedLogger("generate", "storage")
)

// SliverConfigByName - Get a sliver's config by it's codename
func SliverConfigByName(name string) (*SliverConfig, error) {
	bucket, err := db.GetBucket(sliverBucketName)
	if err != nil {
		return nil, err
	}
	rawConfig, err := bucket.Get(fmt.Sprintf("%s.%s", sliverConfigNamespace, name))
	if err != nil {
		return nil, err
	}
	config := &SliverConfig{}
	err = json.Unmarshal(rawConfig, config)
	return config, err
}

// SliverConfigSave - Save a configuration to the database
func SliverConfigSave(config *SliverConfig) error {
	bucket, err := db.GetBucket(sliverBucketName)
	if err != nil {
		return err
	}
	rawConfig, err := json.Marshal(config)
	if err != nil {
		return err
	}
	storageLog.Infof("Saved config for '%s'", config.Name)
	return bucket.Set(fmt.Sprintf("%s.%s", sliverConfigNamespace, config.Name), rawConfig)
}

// SliverFileSave - Saves a binary file into the database
func SliverFileSave(name, fpath string) error {
	bucket, err := db.GetBucket(sliverBucketName)
	if err != nil {
		return err
	}

	rootAppDir, _ := filepath.Abs(assets.GetRootAppDir())
	fpath, _ = filepath.Abs(fpath)
	if !strings.HasPrefix(fpath, rootAppDir) {
		return fmt.Errorf("Invalid path '%s' is not a subdirectory of '%s'", fpath, rootAppDir)
	}

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	storageLog.Infof("Saved '%s' file to database %d byte(s)", name, len(data))
	bucket.Set(fmt.Sprintf("%s.%s", sliverDatetimeNamespace, name), []byte(time.Now().Format(time.RFC1123)))
	return bucket.Set(fmt.Sprintf("%s.%s", sliverFileNamespace, name), data)
}

// SliverFileByName - Saves a binary file into the database
func SliverFileByName(name string) ([]byte, error) {
	bucket, err := db.GetBucket(sliverBucketName)
	if err != nil {
		return nil, err
	}
	return bucket.Get(fmt.Sprintf("%s.%s", sliverFileNamespace, name))
}

// SliverFiles - List all sliver files
func SliverFiles() ([]string, error) {
	bucket, err := db.GetBucket(sliverBucketName)
	if err != nil {
		return nil, err
	}
	keys, err := bucket.List(sliverFileNamespace)
	if err != nil {
		return nil, err
	}

	// Remove namespace prefix
	names := []string{}
	for _, key := range keys {
		names = append(names, key[len(sliverFileNamespace):])
	}
	return names, nil
}