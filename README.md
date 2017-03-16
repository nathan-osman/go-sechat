## go-sechat

[![GoDoc](https://godoc.org/github.com/nathan-osman/go-sechat?status.svg)](https://godoc.org/github.com/nathan-osman/go-sechat)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

The [Stack Exchange chat network](http://chat.stackexchange.com) does not provide an official API. This package aims to bridge that gap for Go applications by providing a simple interface with native Go primitives (channels, etc.).

### Features

As go-sechat grows, more and more chat functionality is exposed through the package. Here's a list of just some of the things go-sechat can do:

- Run completely headless with no UI or embedded browser
- Login using Stack Exchange credentials
- Maintain a persistent connection to the chat server; reauthenticating, pausing, and reconnecting when a failure occurs
- Join, create, leave, and invite users to rooms
- Perform basic chat activities, such as posting and starring messages
- Receive a stream of all events from rooms that have been joined
- Intelligently retry failed messages when throttling occurs
- Upload images to `i.stack.imgur.com`

### Usage

To use the package in an application, simply import it:

    import "github.com/nathan-osman/go-sechat"

Examples are provided in the [package documentation](https://godoc.org/github.com/nathan-osman/go-sechat).
