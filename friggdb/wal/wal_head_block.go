package wal

import (
	"bytes"
	"os"
	"sort"

	bloom "github.com/dgraph-io/ristretto/z"
	"github.com/dgryski/go-farm"
	"github.com/google/uuid"
	"github.com/grafana/frigg/friggdb/encoding"
)

type HeadBlock interface {
	CompleteBlock

	Write(id encoding.ID, b []byte) error
	Complete(w WAL) (CompleteBlock, error)
	Length() int
}

type headBlock struct {
	completeBlock

	appendFile *os.File
}

func (h *headBlock) Write(id encoding.ID, b []byte) error {
	start, length, err := h.appendObject(id, b)
	if err != nil {
		return err
	}

	// insert sorted to records
	i := sort.Search(len(h.records), func(idx int) bool {
		return bytes.Compare(h.records[idx].ID, id) == 1
	})
	h.records = append(h.records, nil)
	copy(h.records[i+1:], h.records[i:])
	h.records[i] = &encoding.Record{
		ID:     id,
		Start:  start,
		Length: length,
	}

	h.meta.ObjectAdded(id)
	return nil
}

func (h *headBlock) Length() int {
	return len(h.records)
}

func (h *headBlock) Complete(w WAL) (CompleteBlock, error) {
	if h.appendFile != nil {
		err := h.appendFile.Close()
		if err != nil {
			return nil, err
		}
	}

	walConfig := w.config()

	// 1) create a new block in work dir
	// 2) append all objects from this block in order
	// 3) move from workdir -> realdir
	// 4) remove old
	orderedBlock := &headBlock{
		completeBlock: completeBlock{
			meta:     encoding.NewBlockMeta(h.meta.TenantID, uuid.New()),
			filepath: walConfig.WorkFilepath,
			records:  make([]*encoding.Record, 0, len(h.records)/walConfig.IndexDownsample+1),
			bloom:    bloom.NewBloomFilter(float64(len(h.records)), walConfig.BloomFP),
		},
	}
	orderedBlock.meta.StartTime = h.meta.StartTime
	orderedBlock.meta.EndTime = h.meta.EndTime
	orderedBlock.meta.MinID = h.meta.MinID
	orderedBlock.meta.MaxID = h.meta.MaxID

	_, err := os.Create(orderedBlock.fullFilename())
	if err != nil {
		return nil, err
	}

	// records are already sorted
	var currentRecord *encoding.Record
	for i, r := range h.records {
		b, err := h.readRecordBytes(r)
		if err != nil {
			return nil, err
		}

		start, length, err := orderedBlock.appendBytes(b)
		if err != nil {
			return nil, err
		}

		orderedBlock.bloom.Add(farm.Fingerprint64(r.ID))

		// start or continue working on a record
		if currentRecord == nil {
			currentRecord = &encoding.Record{
				ID:     r.ID,
				Start:  start,
				Length: length,
			}
		} else {
			currentRecord.Length += length
		}

		// if this is the last record to be included by hte downsample config OR is simply the last record
		if i%walConfig.IndexDownsample == walConfig.IndexDownsample-1 ||
			i == len(h.records)-1 {
			currentRecord.ID = r.ID
			orderedBlock.records = append(orderedBlock.records, currentRecord)
			currentRecord = nil
		}
	}

	workFilename := orderedBlock.fullFilename()
	orderedBlock.filepath = h.filepath
	completeFilename := orderedBlock.fullFilename()

	err = os.Rename(workFilename, completeFilename)
	if err != nil {
		return nil, err
	}

	os.Remove(h.fullFilename())
	if err != nil {
		return nil, err
	}

	return orderedBlock, nil
}

func (h *headBlock) appendObject(id encoding.ID, b []byte) (uint64, uint32, error) {
	if h.appendFile == nil {
		name := h.fullFilename()

		f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, 0, err
		}
		h.appendFile = f
	}

	info, err := h.appendFile.Stat()
	if err != nil {
		return 0, 0, err
	}

	length, err := encoding.MarshalObjectToWriter(id, b, h.appendFile)
	if err != nil {
		return 0, 0, err
	}

	return uint64(info.Size()), uint32(length), nil
}

func (h *headBlock) appendBytes(b []byte) (uint64, uint32, error) {
	if h.appendFile == nil {
		name := h.fullFilename()

		f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, 0, err
		}
		h.appendFile = f
	}

	info, err := h.appendFile.Stat()
	if err != nil {
		return 0, 0, err
	}

	_, err = h.appendFile.Write(b)
	if err != nil {
		return 0, 0, err
	}

	return uint64(info.Size()), uint32(len(b)), nil
}