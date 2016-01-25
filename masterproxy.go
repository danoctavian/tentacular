package main

import (
  "github.com/elazarl/goproxy"
  "net/http"
  "net/url"
  "sync"
)

type MasterProxy struct {
  slaveProxies *SlaveProxies

}

func NewMasterProxy(slaveProxies *SlaveProxies) *MasterProxy {
  master := MasterProxy{slaveProxies: slaveProxies}
  return &master
}

/* handle an incoming request */
func (p *MasterProxy) OnRequest(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {
  return nil, nil
}

/* handle a request on its way out to be proxied */
func (p* MasterProxy) Proxy(*http.Request) (*url.URL, error) {

  return nil, nil
}


type MasterProxyConfig struct {
  slaveProxyPort uint16
}


/* keeps track of current slave proxies */
type SlaveProxies struct {

  slaves map[SlaveContact]bool
  mutex *sync.Mutex
}

func NewSlaveProxies(port uint16) *SlaveProxies {
  return nil
}

func (ps *SlaveProxies) Run() {

}

func (ps *SlaveProxies) GetURLs() []url.URL {
  urls := []url.URL{}
  for slave := range ps.slaves {
    urls = append(urls, slave.url)
  }
  return urls
}

type SlaveContact struct {
  url url.URL
}

func NewMasterProxyServer(config MasterProxyConfig) *goproxy.ProxyHttpServer {

  slaveProxies := NewSlaveProxies(config.slaveProxyPort)

  // launch handling of slave proxies
  go slaveProxies.Run()

  masterProxy := NewMasterProxy(slaveProxies)

  proxy := goproxy.NewProxyHttpServer()
  proxy.Verbose = true

  transport := http.Transport{Proxy: masterProxy.Proxy}
  proxy.Tr = transport
  proxy.OnRequest().DoFunc(masterProxy.OnRequest)

  // make it do it's work in the background

  return proxy
}