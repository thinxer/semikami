semikami [![Build Status](https://travis-ci.org/thinxer/semikami.svg)](https://travis-ci.org/thinxer/semikami) [![GoDoc](https://godoc.org/gopkg.in/thinxer/semikami.v2?status.svg)](https://godoc.org/gopkg.in/thinxer/semikami.v2)
========

[chi](https://github.com/pressly/chi) is a better alternative since `context.Context` is now part of `http.Request`. It is more feature complete. The only advantage `semikami` gets here, is less allocations, as `Request.WithConext` does a shallow copy of the request.

A simple middleware framework for Go, inspired by [kami](https://github.com/guregu/kami).

install
=======
```
go get gopkg.in/thinxer/semikami.v2
```

