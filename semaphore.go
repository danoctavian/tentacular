package main

type empty interface{}
type Semaphore chan empty

// acquire n resources
func (s Semaphore) Acquire(n int) {
  e := 1
  for i := 0; i < n; i++ {
    s <- e
  }
}

// release n resources
func (s Semaphore) Release(n int) {
  for i := 0; i < n; i++ {
    <-s
  }
}
