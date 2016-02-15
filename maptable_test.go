package main

import (
  "testing"
  "strconv"
)

func TestMapOperations(t *testing.T) {
  mt := NewMapTable(100)


  k1 := "k1"
  v1 := "v1"

  k2 := "k2"
  v2 := "v2"

  mt.Add(HashString(k1), k1, v1)

  for i := 3; i < 10; i++ {
    k := "k" + strconv.Itoa(i)
    v := "v" + strconv.Itoa(i)
    mt.Add(HashString(k), k, v)
  }

  mt.Add(HashString(k2), k2, v2)

  getV2 := mt.Get(HashString(k2), k2)
  getV1 := mt.Get(HashString(k1), k1)
  if getV2 != v2 {
    t.Errorf("get failed")
    return
  }

  if getV1 != v1 {
    t.Errorf("get failed")
    return
  }

  k4 := "k4"

  mt.Delete(HashString(k4), k4)

  if mt.Get(HashString(k4), k4) != nil {
    t.Errorf("delete failed")
  }
}
