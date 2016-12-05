## go-sechat

[![GoDoc](https://godoc.org/github.com/nathan-osman/go-sechat?status.svg)](https://godoc.org/github.com/nathan-osman/go-sechat)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

The [Stack Exchange chat network](http://chat.stackexchange.com) does not provide an official API. This package aims to bridge that gap for Go applications by providing a simple interface with native Go primitives (channels, etc.).

### Usage

To use the package, simply import it:

    import "github.com/nathan-osman/go-sechat"

### Authentication

In order to make authenticated requests, create an `Auth` object and invoke its `Login()` method:

    a, _ := sechat.NewAuth("email", "password")
    if err := a.Login(); err != nil {
        // process login error
    }

Once logged in, the `Save()` method can be used to obtain a serializable object. This can later be passed to `Load()` to restore the authentication data:

    s, _ := a.Save()
    // ...
    // do something with s, such as saving to a file
    // ...
    a.Load(s)

### Posting Messages

To post a message, create an instance of `Conn` and invoke `Send()`:

    c, _ := sechat.NewConn(auth)
    c.Send(201, "Testing go-sechat...")

`NewConn` expects a valid `Auth` to be passed as the first parameter. `Send()` expects the room ID and the actual message to send.
