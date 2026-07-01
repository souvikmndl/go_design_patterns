package adapter

/*
A classic, practical example of the Adapter Pattern is integrating a modern system that uses JSON with a legacy system (or a third-party analytics tool) that only understands XML.

The Adapter acts as a translator, allowing these incompatible interfaces to work together seamlessly without modifying the original codebases.

The Scenario
Imagine we have a modern Go application that processes user data in JSON format. However, we need to send this data to a legacy XML Analytics Service.

Here is how we implement the Adapter pattern to bridge them.
*/

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// 1. The Target Interface (What our modern client expects)
type JSONAnalyticalTarget interface {
	SendJSONData(jsonBytes []byte) string
}

// 2. The Adaptee (The legacy/third-party system with an incompatible interface)
type LegacyXMLAnalyticsService struct{}

func (l *LegacyXMLAnalyticsService) SendXMLData(xmlBytes []byte) string {
	return fmt.Sprintf("Legacy Service received XML: %s", string(xmlBytes))
}

// 3. The Helper Structs for Data Conversion
type UserData struct {
	ID   int    `json:"id" xml:"id"`
	Name string `json:"name" xml:"name"`
}

// 4. The Adapter (Bridges JSONAnalyticalTarget and LegacyXMLAnalyticsService)
type JSONToXMLAdapter struct {
	LegacyService *LegacyXMLAnalyticsService
}

func (a *JSONToXMLAdapter) SendJSONData(jsonBytes []byte) string {
	// Step A: Unmarshal the incoming JSON data into a Go struct
	var data UserData
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Sprintf("Adapter Error: failed to unmarshal JSON: %v", err)
	}

	// Step B: Marshal that struct into XML format required by the legacy service
	xmlBytes, err := xml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Adapter Error: failed to marshal XML: %v", err)
	}

	// Step C: Delegate the actual work to the Adaptee using the converted format
	return a.LegacyService.SendXMLData(xmlBytes)
}

// 5. The Client Code
func JSONToXML() {
	// Our client has data in JSON format
	inputJSON := []byte(`{"id": 42, "name": "Alice"}`)

	// We instantiate the legacy system
	legacyService := &LegacyXMLAnalyticsService{}

	// We wrap the legacy system inside our Adapter
	var adapter JSONAnalyticalTarget = &JSONToXMLAdapter{
		LegacyService: legacyService,
	}

	// The client talks seamlessly to the Adapter using the JSON interface
	response := adapter.SendJSONData(inputJSON)

	fmt.Println(response)
}
