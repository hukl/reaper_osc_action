package main

import (
	"net"
	"bytes"
	"time"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "net/url"
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

// padString pads a string to a multiple of 4 bytes by adding null bytes (0x00).
func padString(input string) []byte {
	// Start by appending a null terminator to the input string
	strWithNull := input + "\x00"

	// Calculate the total length including the null character
	length := len(strWithNull)

	// Calculate how much padding is required to make the length a multiple of 4
	padding := (4 - (length % 4)) % 4

	// Append the necessary number of null bytes (0-3)
	paddedString := strWithNull + string(bytes.Repeat([]byte{'\x00'}, padding))

	result := []byte(paddedString)
	return result
}

// createOSCPacket constructs the OSC packet with address, type tags, and arguments,
// and then applies padding to make the final packet a multiple of 4 bytes.
func createOSCPacket(address, argument string) []byte {
    var buf bytes.Buffer

    // Write the OSC address (e.g., "/action")
    buf.Write(padString(address))

    // Write the OSC type tag (e.g., ",s" for a string argument)
    buf.Write(padString(",s"))

    // Write the OSC argument (e.g., "_S&M_INS_MARKER_PLAY")
    buf.Write(padString(argument))

    return buf.Bytes()
}

func sendOSC(ip string, port int, commandID string) {
	client, err := net.ListenPacket("udp", "0.0.0.0:")
	if err != nil {
		return
	}

	// It seems to be good practice to close the "connection" when the function
	// is terminating but it seems the socket is closed before the packet is
	// actually sent which is why the timer is added in the end.
	// Maybe this should all happen in a goroutine anyway but not there yet
	defer client.Close()

	packet := createOSCPacket("/action", commandID)

	RemoteAddr := net.UDPAddr{IP: net.ParseIP(ip), Port: port}

	client.WriteTo(packet, &RemoteAddr)

	// Can't get this code to work without the 50ms timer to allow
	// the socket buffer to flush before the client connection is closed
	// Not sure whether I should not defer the client.Close() in this
	// context?
	time.Sleep(50 * time.Millisecond)
}

func handleEvent(event Event) {
    // Extract the IP, Port, and Command ID from the global settings
    ip        := currentSettings.IPAddress
    port      := currentSettings.Port
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

