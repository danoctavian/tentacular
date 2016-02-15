package main
import (
  "sync"
  "hash/fnv"
)


/*
  thread safe map implemented as a fixed size array of maps.
  to avoid having a single contention point
 */
type MapTable struct {
  table []Bucket
}

type Bucket struct {
  contents *map[interface{}] interface{}
  mutex *sync.Mutex
}

func NewMapTable(tableSize int) *MapTable {
  table := make([]Bucket, tableSize)

  // initialize the array  buckets
  for i, _ := range table {
    mp := make(map[interface{}]interface{})
    table[i] = Bucket{&mp,&sync.Mutex{}}
  }

  mapTable := MapTable{table}
  return &mapTable
}

func (mt *MapTable) Add(keyHash uint32, key interface{}, value interface{}) {
  bucket := mt.getBucket(keyHash)

  bucket.mutex.Lock()
  defer bucket.mutex.Unlock()
  (*bucket.contents)[key] = value
}

func (mt *MapTable) Get(keyHash uint32, key interface{}) interface{} {
  bucket := mt.getBucket(keyHash)

  bucket.mutex.Lock()
  defer bucket.mutex.Unlock()
  return (*bucket.contents)[key]
}

func (mt *MapTable) Delete(keyHash uint32, key interface{}) {
  bucket := mt.getBucket(keyHash)

  bucket.mutex.Lock()
  defer bucket.mutex.Unlock()
  delete(*bucket.contents, key)
}

func (mt *MapTable) Has(keyHash uint32, key interface{}) bool {
  bucket := mt.getBucket(keyHash)

  bucket.mutex.Lock()
  defer bucket.mutex.Unlock()
  _, present := (*bucket.contents)[key]
  return present
}

func (mt *MapTable) getBucket(keyHash uint32) Bucket {
  return mt.table[int(keyHash) % len(mt.table)]
}

/* HASHING METHODS */
func HashString(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}
