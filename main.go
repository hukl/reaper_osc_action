package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"reaper_osc_action/osc"

	// External Dependencies
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

// Struct to hold Event data coming from StreamDeck
type Event struct {
	Event   string `json:"event"`
	Action  string `json:"action"`
	Context string `json:"context"`
	Payload struct {
		Settings Settings `json:"settings"`
	} `json:"payload"`
}

// Method to check if any settings have been provided or still null values
func (e Event) hasSettings() bool {
	return e.Payload.Settings.IPAddress != "" || e.Payload.Settings.Port != 0 || e.Payload.Settings.CommandID != ""
}

// Global variable to store settings per StreamDeck context
var settingsMap = make(map[string]Settings)

////////////////////////////////////////////////////////////////////////////////
// Event Handlers                                                             //
////////////////////////////////////////////////////////////////////////////////

func handleEvent(event Event, udp_client net.PacketConn) {
	context := event.Context // Unique context for each plugin instance

	// Retrieve the settings for the current context
	settings, exists := settingsMap[context]
	if !exists {
		log.Printf("No settings found for context: %s\n", context)
		return
	}

	// Extract the IP, Port, and Command ID from the instance-specific settings
	ip := settings.IPAddress
	port := settings.Port
	commandID := settings.CommandID

	log.Printf("Received keyDown event for context %s: Triggering OSC action with IP: %s, Port: %d, Command ID: %s\n", context, ip, port, commandID)
	osc.SendOSC(ip, port, commandID, udp_client)
}

func handleWillAppearEvent(event Event) {
	log.Println("Plugin appeared, initializing settings...")
	context := event.Context // Unique context for each plugin instance

	if event.hasSettings() {
		settingsMap[context] = event.Payload.Settings
		log.Printf("Initialized settings for context %s: IP: %s, Port: %d, Command ID: %s\n", context, settingsMap[context].IPAddress, settingsMap[context].Port, settingsMap[context].CommandID)
	} else {
		log.Println("No settings found for context:", context, "Initializing default settings")
		settingsMap[context] = Settings{
			IPAddress: "127.0.0.1",      // default IP
			Port:      8000,             // default port
			CommandID: "defaultCommand", // default command
		}
	}
}

func handleDidReceiveSettingsEvent(event Event) {
	context := event.Context // Unique context for the plugin instance

	// Update the settings for this context
	log.Printf("Raw settings for context %s: %+v\n", context, event.Payload.Settings)
	settingsMap[context] = event.Payload.Settings // Store the settings for the specific context

	log.Printf("Settings updated for context %s: IP: %q, Port: %d, Command ID: %s\n",
		context,
		settingsMap[context].IPAddress,
		settingsMap[context].Port,
		settingsMap[context].CommandID)
}

////////////////////////////////////////////////////////////////////////////////
// Main Function                                                              //
////////////////////////////////////////////////////////////////////////////////

func main() {
	f, err := os.OpenFile("plugin_debug.log", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

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

	udp_client, err := net.ListenPacket("udp", "0.0.0.0:")
	if err != nil {
		return
	}

	defer udp_client.Close()

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

		// Event Dispatch
		switch event.Event {
		case "keyDown":
			handleEvent(event, udp_client)
		case "willAppear":
			handleWillAppearEvent(event)
		case "didReceiveSettings":
			handleDidReceiveSettingsEvent(event)
		default:
			log.Printf("Unhandled Event Type: %s\n", event.Event)
		}
	}
}
