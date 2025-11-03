package queue

import (
	"io"

	"github.com/nats-io/nats.go"
)

// ObjectStores
const (
	FilesOS = "files"
)

type ObjectStoreNATS struct {
	os nats.ObjectStore
}

var objectStore = map[string]*ObjectStoreNATS{}

// Funcs
func (osNATS *ObjectStoreNATS) Put(
	obj *nats.ObjectMeta,
	reader io.Reader,
	opts ...nats.ObjectOpt,
) (*nats.ObjectInfo, error) {
	return osNATS.os.Put(obj, reader, opts...)
}

func (osNATS *ObjectStoreNATS) Delete(name string) error {
	return osNATS.os.Delete(name)
}

func (osNATS *ObjectStoreNATS) Get(name string) ([]byte, error) {
	return osNATS.os.GetBytes(name)
}

// New ObjectStore
func ObjectStore(bucketName string) *ObjectStoreNATS {
	conn := newConnectionNatsCore()
	// Connect to JetStream
	jsContext, err := conn.JetStream()
	if err != nil {
		panic(err)
	}

	var exists bool
	var bucket *ObjectStoreNATS

	if bucket, exists = objectStore[bucketName]; !exists {
		os, err := jsContext.CreateObjectStore(&nats.ObjectStoreConfig{
			Bucket:   bucketName,
			MaxBytes: 2.5e+7,
			Storage:  nats.FileStorage,
		})
		if err != nil {
			panic(err)
		}
		bucket = &ObjectStoreNATS{
			os: os,
		}
		objectStore[bucketName] = bucket
	}
	return bucket
}
