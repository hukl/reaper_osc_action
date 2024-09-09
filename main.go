package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/url"
    "github.com/hypebeast/go-osc/osc"
    "github.com/gorilla/websocket"
)

// Struct to hold the registration message
type RegistrationMessage struct {
    Event string `json:"event"`
    UUID  string `json:"uuid"`
}

type Settings struct {
    IPAddress string `json:"ipAddress"`
    Port      int    `json:"port"`
    CommandID int    `json:"commandID"`
}

type Event struct {
    Event string `json:"event"`
    Action string `json:"action"`
    Context string `json:"context"`
    Payload struct {
        Settings struct {
            IPAddress string `json:"ipAddress"`
            Port      int    `json:"port"`
            CommandID int    `json:"commandID"`
        } `json:"settings"`
    } `json:"payload"`
}

func sendOSC(ip string, port int, commandID int) {
    // Create OSC client and send the message
    client := osc.NewClient(ip, port)
    msg := osc.NewMessage("/action")
    msg.Append(int32(commandID))

    err := client.Send(msg)
    if err != nil {
        log.Fatalf("Failed to send OSC message: %v", err)
    }
    fmt.Printf("OSC message sent to %s:%d with Command ID: %d\n", ip, port, commandID)
}

func handleEvent(event Event) {
    // Extract the IP, Port, and Command ID from the settings
    ip := event.Payload.Settings.IPAddress
    port := event.Payload.Settings.Port
    commandID := event.Payload.Settings.CommandID

    fmt.Printf("Received event: Triggering OSC action with IP: %s, Port: %d, Command ID: %d\n", ip, port, commandID)
    sendOSC(ip, port, commandID)
}

func main() {
    // Parse command-line arguments passed by StreamDeck
    var port, pluginUUID, registerEvent, info string
    flag.StringVar(&port, "port", "", "WebSocket port provided by StreamDeck")
    flag.StringVar(&pluginUUID, "pluginUUID", "", "Unique plugin UUID")
    flag.StringVar(&registerEvent, "registerEvent", "", "Event type to register the plugin")
    flag.StringVar(&info, "info", "", "Additional StreamDeck information")
    flag.Parse()

    if port == "" || pluginUUID == "" || registerEvent == "" {
        log.Fatal("Required parameters not provided")
        return
    }

    fmt.Printf("Connecting to WebSocket on port: %s\n", port)

    // Establish WebSocket connection
    u := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/"}
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("Failed to connect to WebSocket:", err)
    }
    defer conn.Close()

    // Register the plugin
    registerMessage := RegistrationMessage{
        Event: registerEvent,
        UUID:  pluginUUID,
    }
    jsonMessage, _ := json.Marshal(registerMessage)

    err = conn.WriteMessage(websocket.TextMessage, jsonMessage)
    if err != nil {
        log.Fatal("Failed to register plugin:", err)
    }

    fmt.Println("Plugin registered successfully")

    // Listen for events from StreamDeck
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            return
        }

        // Parse incoming message as JSON
        var event Event
        if err := json.Unmarshal(message, &event); err != nil {
            log.Println("Failed to unmarshal JSON:", err)
            continue
        }

        // Handle the event (trigger the OSC action)
        if event.Event == "keyDown" {  // Handle key down (button press) events
            handleEvent(event)
        }

        if event.Event == "didReceiveSettings" {
            settings := event.Payload.Settings
            fmt.Printf("Settings received: IP: %s, Port: %d, Command ID: %d\n", settings.IPAddress, settings.Port, settings.CommandID)
        }
    }
}
