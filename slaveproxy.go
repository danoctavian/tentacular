package main
import (
  "net/url"
  "github.com/elazarl/goproxy"
  "os"
  "os/signal"
  "syscall"
  "log"
  "net/http"
  "time"
  "strconv"
)

type SlaveProxyConfig struct {
  port uint
  masterURL url.URL
}

const SLAVE_HEARTBEAT_INTERVAL = 500 * time.Millisecond

func NewSlaveProxyServer(config SlaveProxyConfig) *goproxy.ProxyHttpServer {

  portStr := strconv.Itoa(int(config.port))
  // this is rather unnecessary mostly because the master does a cleanup.
  // TODO: consider form removal

  // listen for kill signals to be able to tell the master you are shutting down
  ch := make(chan os.Signal, 1)
  signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
  go func() {
    signalType := <-ch
    signal.Stop(ch)

    log.Println("Exit command received. Notifying master of death.")

    _, err := http.Get(config.masterURL.String() + "/leave?port=" + portStr)
    if err != nil {
      log.Println("No master available at " + config.masterURL.String())
    }

    log.Println("Signal type : ", signalType)
    os.Exit(0)
  }()

  go func() {
    //heartbeat the master
    for {
      time.Sleep(SLAVE_CLEANUP_INTERVAL)
      _, err := http.Get(config.masterURL.String() + "/join?port=" + portStr)
      if err != nil {
        log.Println("No master available at " + config.masterURL.String())
      }
    }
  }()

  proxy := goproxy.NewProxyHttpServer()

  proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
    return goproxy.OkConnect, host
  })

  proxy.Verbose = true
  return proxy
}

