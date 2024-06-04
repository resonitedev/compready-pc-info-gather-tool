package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

func main() {
	fmt.Println("Getting system information from WMI")
	// Query WMI for motherboard information
	var pcParts []PCPart
	var baseboard []Win32Baseboard
	if err := wmi.Query("SELECT * FROM Win32_BaseBoard", &baseboard); err != nil {
		fmt.Println("Error querying Win32_BaseBoard:", err)
		return
	}
	fmt.Println("Got motherboard information")
	for _, board := range baseboard {
		pcParts = append(pcParts, PCPart{
			Type: "Motherboard",
			Info: fmt.Sprintf("%s, %s", board.Manufacturer, board.Product),
		})
	}

	// Query WMI for memory information
	var memory []Win32PhysicalMemory
	if err := wmi.Query("SELECT * FROM Win32_PhysicalMemory", &memory); err != nil {
		fmt.Println("Error querying Win32_PhysicalMemory:", err)
		return
	}
	fmt.Println("Got memory information")
	var (
		manufacturer string
		capacity     uint64
		speed        uint32
		voltage      uint16
		partNumber   string
	)

	for _, mem := range memory {
		if manufacturer != mem.Manufacturer {
			manufacturer = mem.Manufacturer
		}
		if capacity != mem.Capacity {
			capacity = mem.Capacity
		}
		if speed != mem.ConfiguredClockSpeed {
			speed = mem.ConfiguredClockSpeed
		}
		if voltage != mem.ConfiguredVoltage {
			voltage = mem.ConfiguredVoltage
		}
		if partNumber != mem.PartNumber {
			partNumber = mem.PartNumber
		}
	}

	pcParts = append(pcParts, PCPart{
		Type: "RAM",
		Info: fmt.Sprintf("%s %s, %dx%vGB, %d MHz (%v Volts)", manufacturer, strings.TrimSpace(partNumber), len(memory), convertBytesToGB(capacity), speed, convertVolts(float64(voltage))),
	})

	// Query WMI for CPU information
	var cpu []Win32Processor
	if err := wmi.Query("SELECT * FROM Win32_Processor", &cpu); err != nil {
		fmt.Println("Error querying Win32_Processor:", err)
		return
	}
	fmt.Println("Got CPU information")
	for _, processor := range cpu {
		pcParts = append(pcParts, PCPart{
			Type: "CPU",
			Info: fmt.Sprintf(" %s", strings.TrimSpace(processor.Name)),
		})
	}

	// Query WMI for GPU information
	var gpu []Win32VideoController
	if err := wmi.Query("SELECT * FROM Win32_VideoController", &gpu); err != nil {
		fmt.Println("Error querying Win32_VideoController:", err)
		return
	}
	fmt.Println("Got GPU information")
	for _, video := range gpu {
		pcParts = append(pcParts, PCPart{
			Type: "GPU",
			Info: fmt.Sprintf("%s, %s", video.Name, video.DriverVersion),
		})
	}

	maj, _, patch := windows.RtlGetNtVersionNumbers()
	pcParts = append(pcParts, PCPart{
		Type: "Windows Version",
		Info: fmt.Sprintf("%vv%v", maj, patch),
	})

	jsonParts, _ := json.Marshal(pcParts)
	fmt.Println("System JSON Information")
	fmt.Println(string(jsonParts))

	// Write JSON data to file
	file, err := os.Create("system_info.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonParts)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("System JSON Information exported to system_info.json")
	fmt.Println("Press Enter to close the application...")
	fmt.Scanln()
}

func convertBytesToGB(bytes uint64) int {
	gigabytes := float64(bytes) / (1024 * 1024 * 1024)
	return int(gigabytes)
}

func convertVolts(mVolts float64) float64 {
	return mVolts / 1000
}

type Win32Baseboard struct {
	Manufacturer string
	Product      string
}

type Win32PhysicalMemory struct {
	BankLabel            string
	Capacity             uint64
	ConfiguredClockSpeed uint32
	ConfiguredVoltage    uint16
	DeviceLocator        string
	FormFactor           uint16
	InterleaveDataDepth  uint16
	Manufacturer         string
	MaxVoltage           uint16
	MemoryType           uint16
	MinVoltage           uint16
	PartNumber           string
	Speed                uint32
	TotalWidth           uint16
	TypeDetail           uint16
}

type Win32Processor struct {
	Name string
}

type Win32VideoController struct {
	Name          string
	DriverVersion string
}

type PCPart struct {
	Type     string `json:"type"`
	Info     string `json:"info"`
	ImageURL string `json:"image_url,omitempty"`
}
