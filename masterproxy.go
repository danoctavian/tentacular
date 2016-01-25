package main

import (
  "github.com/elazarl/goproxy"
  "net/http"
  "net/url"
  "math/rand"
  "errors"
)

/*

  Current features:

  * load balancing on a set of slave proxies.
  * no timing policies.

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

/* handle an incoming request
   does nothing for now
*/
func (p *MasterProxy) OnRequest(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {
  return r, nil
}

/* handle a request on its way out to be proxied */
func (p* MasterProxy) Proxy(*http.Request) (*url.URL, error) {
  slaves := p.slaveProxies.GetURLs()

  slaveCount := len(slaves)
  if slaveCount == 0 {
    // can also be handled by having the master dispatch by himself
    return nil, errors.New("No slave proxies to dispatch to.")
  }

  /* load balance by randomly distributing */
  chosenSlave := rand.Int() % slaveCount

  return &slaves[chosenSlave], nil
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