package main
import (
  "time"
  "net/url"
  "net/http"
  "log"
  "strconv"
  "sync"
)

/* keeps track of current slave proxies */
type SlaveProxies struct {

  port uint16

  /*  single point of contention - this global lock may cause performance problems

    TODO:
    TRICK to solve it. make an atomic update variable that tells the reader whether
    the URLs have changed and needs to reread them (grabbing the lock) or just refetch
    a cached copy of the last value
  */
  slaves *map[url.URL] *SlaveContact
  mutex *sync.Mutex
}

func NewSlaveProxies(port uint16) *SlaveProxies {

  mp := make(map[url.URL]*SlaveContact)
  return &SlaveProxies{port, &mp, &sync.Mutex{}}
}

func (ps *SlaveProxies) addSlave(url url.URL) {
  ps.mutex.Lock()
  defer ps.mutex.Unlock()
  if val, ok := (*ps.slaves)[url]; ok {
    val.lastSeen = time.Now()
  } else {
    (*ps.slaves)[url] = &SlaveContact{url: url, lastSeen: time.Now()}
  }
}

func (ps *SlaveProxies) removeSlave(url url.URL) {
  ps.mutex.Lock()
  defer ps.mutex.Unlock()
  delete(*ps.slaves, url)
}

func (ps *SlaveProxies) Run() {

  http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
    ps.addSlave(*r.URL)
    w.WriteHeader(200)
  })

  http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
  })

  log.Fatal(http.ListenAndServe(":" + strconv.Itoa(int(ps.port)), nil))
}

func (ps *SlaveProxies) regularCleanup() []url.URL {
  for {
    time.Sleep(SLAVE_CLEANUP_INTERVAL)

    now := time.Now()
    func() {
      ps.mutex.Lock()
      defer ps.mutex.Unlock()

      toBeRemoved := []url.URL{}

      for url, slave := range *ps.slaves {
        if now.Sub(slave.lastSeen).Seconds() >= SLAVE_CLEANUP_INTERVAL.Seconds() {
          toBeRemoved = append(toBeRemoved, url)
        }
      }

      for _, removableUrl := range toBeRemoved {
        delete(*ps.slaves, removableUrl)
      }
    }()
  }
}

func (ps *SlaveProxies) GetURLs() []url.URL {
  ps.mutex.Lock()
  defer ps.mutex.Unlock()

  urls := []url.URL{}
  for url := range *ps.slaves {
    urls = append(urls, url)
  }
  return urls
}

type SlaveContact struct {
  url url.URL
  lastSeen time.Time
}

const SLAVE_CLEANUP_INTERVAL = 1000 * time.Millisecond // milliseconds
