/*
Package go-sechat aims to bridge the gap between the Stack Exchange chat network and Go applications by providing a simple interface with native Go primitives.

Basic Usage

In order to make requests, create a `Conn` object:

    c, err := sechat.New("email@example.com", "passw0rd", 1)
    if err != nil {
        // handle error
    }
    defer c.Close()

Since authentication and connection are done asynchronously, waiting for them to complete is highly recommended:

    if c.WaitForConnected() {
        // do stuff
    }

Interacting with Rooms

To join an additional room, use the `Join()` method:

    // Join the Ask Ubuntu General Room (#201)
    if err := c.Join(201); err != nil {
        // handle error
    }

To leave, use the (appropriately named) `Leave()` method:

    if err := c.Leave(201); err != nil {
        // handle error
    }

To obtain a list of users in the room, use `UsersInRoom()`:

    if users, err := c.UsersInRoom(201); err == nil {
        for _, u := range users {
            fmt.Printf("User: %s\n", u.Name)
        }
    }

To obtain a list of rooms that a user is in, use `User()`:

    if user, err := c.User(1345); err == nil {
        for _, r := range user.Rooms {
            log.Printf("Room: %s\n", r.Name)
        }
    }

The `NewRoom()` method can be used to create new rooms:

    r, err := c.NewRoom(
        "Room Name",
        "Room description",
        "askubuntu.com",        // room host
        sechat.AccessReadWrite, // access
    )
    if err != nil {
        // handle error
    }

In the example above, `r` is an `int` containing the ID of the new room that was created.

Receiving Events

To receive events from the chat server, simply receive from the `Events` channel in `Conn`:

    for e := range c.Events {
        // e is of type *Event
    }

Interacting with Messages

To post a message, simply invoke `Send()`:

    if err := c.Send(201, "Testing go-sechat..."); err != nil {
        // handle error
    }

If the message is in response to an earlier event, the `Reply()` method is also available:

    if err := c.Reply(e, "Reply to event"); err != nil {
        // handle error
    }

Uploading Images

To upload an image, prepare an `io.Reader` and pass it to `Image()`:

    f, err := os.Open("/path/to/image.png")
    if err != nil {
        // handle error
    }
    defer f.Close()
    u, err := c.Image(f)
    if err != nil {
        // handle error
    }

In the example above, `u` is the URL of the newly uploaded image.
*/
package sechat
