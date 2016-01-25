package main
import (
  "net/url"
  "github.com/elazarl/goproxy"
  "os"
  "os/signal"
  "syscall"
  "log"
  "net/http"
)

type SlaveProxyConfig struct {
  masterURL url.URL
}

func NewSlaveProxyServer(config SlaveProxyConfig) *goproxy.ProxyHttpServer {

  // listen for kill signals to be able to tell the master you are shutting down
  ch := make(chan os.Signal, 1)
  signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
  go func() {
    signalType := <-ch
    signal.Stop(ch)

    log.Println("Exit command received. Notifying master of death.")

    _, err := http.Get(config.masterURL.String() + "/leave")
    if err != nil {
      log.Println("No master available at " + config.masterURL.String())
    }

    log.Println("Signal type : ", signalType)
    os.Exit(0)
  }()

  go func() {
    //heartbeat the master
    for {
      _, err := http.Get(config.masterURL.String() + "/join")
      if err != nil {
        log.Println("No master available at " + config.masterURL.String())
      }
    }
  }()

  proxy := goproxy.NewProxyHttpServer()
  proxy.Verbose = true
  return proxy
}

