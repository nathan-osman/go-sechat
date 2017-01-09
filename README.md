## go-sechat

[![GoDoc](https://godoc.org/github.com/nathan-osman/go-sechat?status.svg)](https://godoc.org/github.com/nathan-osman/go-sechat)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

The [Stack Exchange chat network](http://chat.stackexchange.com) does not provide an official API. This package aims to bridge that gap for Go applications by providing a simple interface with native Go primitives (channels, etc.).

### Usage

To use the package, begin by importing it:

    import "github.com/nathan-osman/go-sechat"

In order to make requests, create a `Conn` object and invoke its `Connect()` method:

    c, err := sechat.New("email@example.com", "passw0rd", 1)
    if err != nil {
        // handle error
    }
    if err := c.Connect(); err != nil {
        // handle error
    }
    defer c.Close()

### Posting Messages

To post a message, simply invoke `Send()`:

    if err := c.Send(201, "Testing go-sechat..."); err != nil {
        // handle error
    }
