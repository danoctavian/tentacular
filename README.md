# Tentacular

web scraping framework

goals:

* multiplex http requests over a set of dynamically changing IPs
* be able to add instances at runtime. deal with them disconnecting
* allow for scraping policies: request rate per ip/globally, sticky url paths (bla.com/orgs/X goes to the same machine for all X)
* backoff in case of blocking
* simple AWS deploy script
* ? automatic IP changing (kill instances, spawn others).
