package main

import (
  "github.com/elazarl/goproxy"
  "net/http"
  "net/url"
  "math/rand"
  "errors"
  "net"
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

  domainSemaphores *MapTable
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


// FIXME: there is no cleanup for the domain semaphore once created. this can become a problem
func (p *MasterProxy) applyRequestLimitOnRequest(r *http.Request) {
  if p.hasConcurrentRequestLimit() {
    keyHash, key := addrKeyHash(r.RemoteAddr)
    domainSemaphore := p.domainSemaphores.Get(keyHash, key).(Semaphore)
    if domainSemaphore != nil {
      domainSemaphore.Acquire(1)
    } else {
      semaphore := make(Semaphore, *p.throttleConfig.MaxConcurrentRequestsPerDomain)
      p.domainSemaphores.Add(keyHash, key, semaphore)
    }
  }
}

func (p *MasterProxy) applyRequestLimitOnResponse(r *http.Response) {
  if p.hasConcurrentRequestLimit() {
    keyHash, key := addrKeyHash(r.Request.RemoteAddr)
    domainSemaphore := p.domainSemaphores.Get(keyHash, key).(Semaphore)
    if domainSemaphore != nil {
      domainSemaphore.Release(1)
    } else {
      panic("Cannot have a response before a request.")
    }
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
  if masterProxy.hasConcurrentRequestLimit() {
    mapTable := NewMapTable(1000)
    masterProxy.domainSemaphores = mapTable
  }

  proxy := goproxy.NewProxyHttpServer()

  transport := &http.Transport{Proxy: masterProxy.Proxy}
  proxy.Tr = transport
  proxy.OnRequest().DoFunc(masterProxy.OnRequest)
  proxy.OnResponse().DoFunc(masterProxy.OnResponse)

  return proxy
}

func (p* MasterProxy) hasConcurrentRequestLimit() bool {
  return p.throttleConfig.MaxConcurrentRequestsPerDomain != nil
}

func addrKeyHash(remoteAddr string) (uint32, string) {
  host, _, _ := net.SplitHostPort(remoteAddr)
  return HashString(host), host
}