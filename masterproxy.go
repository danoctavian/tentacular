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
  throttleConfig ThrottleConfig
}

type Slaves interface {
  GetAddresses() []string
}


/* handle an incoming request
   does nothing for now
*/
func (p *MasterProxy) OnRequest(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {
  return r, nil
}

func (p *MasterProxy) OnResponse(r *http.Response, ctx *goproxy.ProxyCtx) (*http.Response) {
  return r
}

func (p *MasterProxy) applyRequestLimitOnRequest(r *http.Request) {
  if p.throttleConfig.MaxConcurrentRequestsPerDomain != nil {
  }
}

func (p *MasterProxy) applyRequestLimitOnResponse(r *http.Response) {
  if p.throttleConfig.MaxConcurrentRequestsPerDomain != nil {

  }
}

/* handle a request on its way out to be proxied */
func (p* MasterProxy) Proxy(*http.Request) (*url.URL, error) {
  slaves := p.slaveProxies.GetAddresses()

  slaveCount := len(slaves)
  if slaveCount == 0 {
    // can also be handled by having the master dispatch by himself
    return nil, errors.New("No slave proxies to dispatch to.")
  }

  /* load balance by randomly distributing */
  chosenSlave, err := url.Parse("http://" + slaves[rand.Int() % slaveCount])

  return chosenSlave, err
}

type MasterProxyConfig struct {
  slaveProxyPort uint16
  throttleConfig ThrottleConfig
}

type ThrottleConfig struct {
  MaxConcurrentRequestsPerDomain *int
}


func NewMasterProxyServer(config MasterProxyConfig) *goproxy.ProxyHttpServer {

  slaveProxies := NewSlaveProxies(config.slaveProxyPort)

  // launch handling of slave proxies
  go slaveProxies.Run()

  masterProxy := MasterProxy{slaveProxies: slaveProxies, throttleConfig: config.throttleConfig}

  proxy := goproxy.NewProxyHttpServer()

  transport := &http.Transport{Proxy: masterProxy.Proxy}
  proxy.Tr = transport
  proxy.OnRequest().DoFunc(masterProxy.OnRequest)
  proxy.OnResponse().DoFunc(masterProxy.OnResponse)

  return proxy
}