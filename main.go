package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
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
    IPAddress string `json:"ip"`
    Port      int    `json:"port"`
    CommandID string `json:"command_id"`
}

type Event struct {
    Event string `json:"event"`
    Action string `json:"action"`
    Context string `json:"context"`
    Payload struct {
        Settings Settings `json:"settings"`
    } `json:"payload"`
}

// Global variable to store settings
var currentSettings Settings

func sendOSC(ip string, port int, commandID string) {
    // Create OSC client and send the message

    log.Printf("COMMAND ID %s", commandID)

    client := osc.NewClient(ip, port)
    msg := osc.NewMessage("/action")
    msg.Append(commandID)

    err := client.Send(msg)
    if err != nil {
        log.Fatalf("Failed to send OSC message: %v", err)
    }
    fmt.Printf("OSC message sent to %s:%d with Command ID: %s\n", ip, port, commandID)
}

func handleEvent(event Event) {
    // Extract the IP, Port, and Command ID from the global settings
    ip := currentSettings.IPAddress
    port := currentSettings.Port
    commandID := currentSettings.CommandID

    log.Printf("Received keyDown event: Triggering OSC action with IP: %s, Port: %d, Command ID: %s\n", ip, port, commandID)
    sendOSC(ip, port, commandID)
}

func handleWillAppearEvent(event Event) {
    // If no settings are present, initialize default settings
    if event.Payload.Settings.IPAddress == "" || event.Payload.Settings.CommandID == "" {
        log.Println("No settings found, initializing default settings")
        currentSettings = Settings{
            IPAddress: "127.0.0.1", // default IP
            Port:      8000,        // default port
            CommandID: "defaultCommand", // default command
        }
        log.Printf("Initialized settings: IP: %s, Port: %d, Command ID: %s\n", currentSettings.IPAddress, currentSettings.Port, currentSettings.CommandID)
    } else {
        // If settings are present, load them
        log.Printf("Settings loaded from willAppear: IP: %s, Port: %d, Command ID: %s\n", event.Payload.Settings.IPAddress, event.Payload.Settings.Port, event.Payload.Settings.CommandID)
        currentSettings = event.Payload.Settings
    }
}

func main() {
    f, err := os.OpenFile("plugin_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    log.SetOutput(f)

    log.Println("This is a debug message written to plugin_debug.log")




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

        // Handle the keyDown event (button press)
        if event.Event == "keyDown" {
            handleEvent(event)
        }

        if event.Event == "willAppear" {
            log.Println("Plugin appeared, initializing settings...")
            handleWillAppearEvent(event)
        }

        // Handle the didReceiveSettings event to update the global settings
        if event.Event == "didReceiveSettings" {
            log.Printf("Raw settings: %+v\n", event.Payload.Settings)
            currentSettings = event.Payload.Settings
            log.Printf("Settings received: IP: %q, Port: %d, Command ID: %s\n", currentSettings.IPAddress, currentSettings.Port, currentSettings.CommandID)
        }
    }
}

