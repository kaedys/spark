# Spark
Golang client for the Cisco Spark API

## Supported API calls
### Rooms
Method | Description
--- | ---
GetRoom | Gets a room's details by ID
GetRoomByName | Gets the first room that matches the provided name
ListRooms | Lists accessible rooms
CreateRoom | Creates a new room
UpdateRoomName | Updates a room's name
DeleteRoom | Deletes a room by ID

### Messages
Method | Description
--- | --- 
GetMessage | Gets a message by ID
ListMessages | Lists messages in a room
CreateMessage | Sends a new message to a room or directly to person
DeleteMessage | Deletes a message by ID

### Person
Method | Description
--- | --- 
GetPerson | Gets a person's details by ID
ListPeople | Lists existing people (non-admins require email or display name)
CreatePerson | Creates a new person (admin only) 
UpdatePerson | Updates an existing person by ID (admin only) 
DeletePerson | Deletes an existing person by ID (admin only) 

### Webhooks
Method | Description
--- | --- 
GetWebhook | Gets a webhook's details by ID
ListWebhooks | Lists existing webhooks
CreateWebhook | Creates a new webhook
UpdateWebhook | Updates an existing webhook by ID
DeleteWebhook | Deletes an existing webhook by ID 

## Example
```go
package main

import (
    "github.com/kaedys/spark"
)

const (
    token    = "your-spark-access-id"
    roomName = "your-spark-room-name"
)

func main() {
    s := spark.New(token)
    
    // Get the room ID by name
    room, err := s.GetRoomByName(roomName)
    if err != nil {
    	panic(err)
    }
    
    // Define the message we want to send
    m := &spark.NewMessage{
    	RoomID:   room.ID,
    	Markdown: "# Big Message right here!",
    }
    
    // Post the message to the room
    if _, err := s.CreateMessage(m); err != nil {
        panic(err)
    }
}
```

Based on [vallard/spark](https://github.com/vallard/spark), which was inspired by [bluele/slack](https://github.com/bleule/slack). 
