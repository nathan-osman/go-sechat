## go-sechat

[![GoDoc](https://godoc.org/github.com/nathan-osman/go-sechat?status.svg)](https://godoc.org/github.com/nathan-osman/go-sechat)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

The [Stack Exchange chat network](http://chat.stackexchange.com) does not provide an official API. This package aims to bridge that gap for Go applications by providing a simple interface with native Go primitives (channels, etc.).

### Usage

To use the package, simply import it:

    import "github.com/nathan-osman/go-sechat"

In order to make authenticated requests, create an `Auth` object and invoke its `Login()` method:

    a, _ := sechat.NewAuth(&sechat.AuthState{
        Email: "",
        Password: "",
    })
    if err := a.Login(); err != nil {
        // process login error
    }
