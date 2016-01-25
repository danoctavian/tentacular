package main

import (
  "github.com/elazarl/goproxy"
  "net/http"
  "net/url"
)

/*
  TODO:

  currently the master proxy takes a happy-path approach and does not deal with
  http request timeouts or slaves dying. handle these exceptional cases.

 */
type MasterProxy struct {
  slaveProxies Slaves

}

type Slaves interface {
  GetURLs() []url.URL
}

func NewMasterProxy(slaveProxies Slaves) *MasterProxy {
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

func NewMasterProxyServer(config MasterProxyConfig) *goproxy.ProxyHttpServer {

  slaveProxies := NewSlaveProxies(config.slaveProxyPort)

  // launch handling of slave proxies
  go slaveProxies.Run()

  masterProxy := NewMasterProxy(slaveProxies)

  proxy := goproxy.NewProxyHttpServer()
  proxy.Verbose = true

  transport := &http.Transport{Proxy: masterProxy.Proxy}
  proxy.Tr = transport
  proxy.OnRequest().DoFunc(masterProxy.OnRequest)

  return proxy
}