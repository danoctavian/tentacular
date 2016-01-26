package main
import (
  "time"
  "net/http"
  "log"
  "strconv"
  "sync"
  "net"
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
  slaves *map[string] *SlaveContact
  mutex *sync.Mutex
}

func NewSlaveProxies(port uint16) *SlaveProxies {

  mp := make(map[string]*SlaveContact)
  return &SlaveProxies{port, &mp, &sync.Mutex{}}
}

func (ps *SlaveProxies) addSlave(addr string) {
  log.Println("Adding slave with addr " + addr)
  ps.mutex.Lock()
  defer ps.mutex.Unlock()
  if val, ok := (*ps.slaves)[addr]; ok {
    val.lastSeen = time.Now()
  } else {
    (*ps.slaves)[addr] = &SlaveContact{addr: addr, lastSeen: time.Now()}
  }
}

func (ps *SlaveProxies) removeSlave(addr string) {
  log.Println("Removing slave with addr " + addr)
  ps.mutex.Lock()
  defer ps.mutex.Unlock()
  delete(*ps.slaves, addr)
}

func (ps *SlaveProxies) Run() {

  portStr := strconv.Itoa(int(ps.port))

  http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
    ps.addSlave(remoteSlaveAddress(r))
    w.WriteHeader(200)
  })

  http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
    ps.removeSlave(remoteSlaveAddress(r))
    w.WriteHeader(200)
  })

  log.Fatal(http.ListenAndServe(":" + portStr, nil))
}

func remoteSlaveAddress(r *http.Request) string {
  host, _, _ := net.SplitHostPort(r.RemoteAddr)
  port := r.URL.Query().Get("port") // get the advertised port
  return host + ":" + port
}

func (ps *SlaveProxies) regularCleanup() {
  for {
    time.Sleep(SLAVE_CLEANUP_INTERVAL)

    now := time.Now()
    func() {
      ps.mutex.Lock()
      defer ps.mutex.Unlock()

      toBeRemoved := []string{}

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

func (ps *SlaveProxies) GetAddresses() []string {
  ps.mutex.Lock()
  defer ps.mutex.Unlock()

  urls := []string{}
  for url := range *ps.slaves {
    urls = append(urls, url)
  }
  return urls
}

type SlaveContact struct {
  addr string
  lastSeen time.Time
}

const SLAVE_CLEANUP_INTERVAL = 1000 * time.Millisecond // milliseconds
