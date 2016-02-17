package main

import (
  "log"
  "flag"
  "github.com/elazarl/goproxy"
  "net/http"
  "net/url"
  "os"
  "strconv"
)


const MASTER = "master"
const SLAVE = "slave"

func main() {
  proxyType := flag.String("type", SLAVE, "type of proxy. can be either master or slave")

  port := flag.Uint("port", 8080, "port of the proxy")

  slaveport := flag.Uint("slaveport", 6666, "port for slaves to contact")

  urlString := flag.String("masterurl", "http://127.0.0.1:6666", "url of the master proxy")

  maxRequests := flag.Int("maxConcReq", -1, "max concurrent requests to a single domain.")

  flag.Parse()

  var proxyServer *goproxy.ProxyHttpServer = nil

  switch (*proxyType) {
  case MASTER:
    log.Print("Launching a master proxy")
    if *maxRequests == -1 {
      maxRequests = nil
    } else {
      log.Print("applying max concurrent requests per domain limit of %i", maxRequests)
    }
    throttle := ThrottleConfig{MaxConcurrentRequestsPerDomain: maxRequests}
    proxyServer = NewMasterProxyServer(MasterProxyConfig{slaveProxyPort: uint16(*slaveport), throttleConfig: throttle})
  case SLAVE:
    url, err := url.Parse(*urlString)

    if err != nil {
      log.Fatalln("Failed to parse master URL")
      os.Exit(1)
    }

    log.Print("Launching a slave proxy with a master at " + *urlString)
    proxyServer = NewSlaveProxyServer(SlaveProxyConfig{port: *port, masterURL: *url})
  }

  proxyServer.Verbose = false

  log.Println("... on port " + strconv.Itoa(int(*port)))

  log.Fatal(http.ListenAndServe(":" + strconv.Itoa(int(*port)), proxyServer))
}
