package wal

import (
	"os"
	"time"

	bloom "github.com/dgraph-io/ristretto/z"
	"github.com/google/uuid"
	"github.com/grafana/frigg/friggdb/backend"
)

// complete block has all of the fields jpe - make this comment not suck
type completeBlock struct {
	block

	bloom       *bloom.Bloom
	records     []*backend.Record
	timeWritten time.Time
}

type ReplayBlock interface {
	Iterator() (backend.Iterator, error)
	TenantID() string
	Clear() error
}

type CompleteBlock interface {
	WriteableBlock
	ReplayBlock

	Find(id backend.ID) ([]byte, error)
	TimeWritten() time.Time
}

type WriteableBlock interface {
	BlockMeta() *backend.BlockMeta
	BloomFilter() *bloom.Bloom
	BlockWroteSuccessfully(t time.Time)
	WriteInfo() (blockID uuid.UUID, tenantID string, records []*backend.Record, filepath string) // todo:  i hate this method.  do something better. jpe. this goes now
}

func (c *completeBlock) TenantID() string {
	return c.meta.TenantID
}

func (c *completeBlock) WriteInfo() (uuid.UUID, string, []*backend.Record, string) {
	return c.meta.BlockID, c.meta.TenantID, c.records, c.fullFilename()
}

func (c *completeBlock) Find(id backend.ID) ([]byte, error) {
	file, err := c.file()
	if err != nil {
		return nil, err
	}

	finder := backend.NewFinder(c.records, file)

	return finder.Find(id)
}

func (c *completeBlock) Iterator() (backend.Iterator, error) {
	name := c.fullFilename()
	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return backend.NewIterator(f), nil
}

func (c *completeBlock) Clear() error {
	if c.readFile != nil {
		err := c.readFile.Close()
		if err != nil {
			return err
		}
	}

	name := c.fullFilename()
	return os.Remove(name)
}

func (c *completeBlock) TimeWritten() time.Time {
	return c.timeWritten
}

func (c *completeBlock) BlockWroteSuccessfully(t time.Time) {
	c.timeWritten = t
}

func (c *completeBlock) BlockMeta() *backend.BlockMeta {
	return c.meta
}

func (c *completeBlock) BloomFilter() *bloom.Bloom {
	return c.bloom
}
