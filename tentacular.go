package main

import (
  "log"
  "flag"
  "github.com/elazarl/goproxy"
  "net/http"
  "net/url"
  "os"
)


const MASTER = "master"
const SLAVE = "slave"

func main() {
  proxyType := flag.String("type", SLAVE, "type of proxy. can be either master or slave")
  port := flag.Uint("port", 8080, "port of the proxy")

  var proxyServer *goproxy.ProxyHttpServer = nil
  switch (*proxyType) {
  case MASTER:
    slaveport := uint16(*flag.Uint("slaveport", 6666, "port for slaves to contact"))

    proxyServer = NewMasterProxyServer(MasterProxyConfig{slaveProxyPort: slaveport})
  case SLAVE:
    urlString := *flag.String("masterurl", "http://localhost:6666", "url of the master proxy")

    url, err := url.Parse(urlString)

    if err != nil {
      log.Fatalln("Failed to parse master URL")
      os.Exit(1)
    }

    proxyServer = NewSlaveProxyServer(SlaveProxyConfig{masterURL: *url})
  }

  proxyServer.Verbose = true

  log.Fatal(http.ListenAndServe(":" + string(*port), proxyServer))
}
