package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

type SystemInfo struct {
	Motherboard     string          `json:"motherboard"`
	Gpu             string          `json:"gpu"`
	Cpu             string          `json:"cpu"`
	Memory          Memory          `json:"memory"`
	OperatingSystem OperatingSystem `json:"operatingSystem"`
}

type Memory struct {
	Type       string    `json:"type"`
	Speed      int       `json:"speed"`
	Capacities [4]string `json:"capacities"`
	Name       string    `json:"name"`
}

type OperatingSystem struct {
	Version string `json:"version"`
	Patch   string `json:"patch"`
}

func main() {
	var systemInfo SystemInfo
	fmt.Println("Getting system information from WMI")
	// Query WMI for motherboard information
	var baseboard []Win32Baseboard
	if err := wmi.Query("SELECT * FROM Win32_BaseBoard", &baseboard); err != nil {
		fmt.Println("Error querying Win32_BaseBoard:", err)
		return
	}
	fmt.Println("Got motherboard information")
	for _, board := range baseboard {
		fmt.Println(fmt.Sprintf("%s, %s", board.Manufacturer, board.Product))
		systemInfo.Motherboard = fmt.Sprintf("%s, %s", board.Manufacturer, board.Product)
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
		speed        uint32
		voltage      uint16
		partNumber   string
		memType      string
	)

	capacities := [4]string{"N/A", "N/A", "N/A", "N/A"}

	for _, mem := range memory {
		fmt.Println(fmt.Sprintf("Bank: %v | Locator: %v | Speed: %v | Voltage: %v | Part Number: %v", mem.BankLabel, mem.DeviceLocator, mem.ConfiguredClockSpeed, mem.ConfiguredClockSpeed, mem.PartNumber))
		switch mem.BankLabel {
		case "P0 CHANNEL A":
			if mem.DeviceLocator == "DIMM 1" {
				capacities[0] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
			} else if mem.DeviceLocator == "DIMM 2" {
				capacities[1] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
			}
		case "P0 CHANNEL B":
			if mem.DeviceLocator == "DIMM 1" {
				capacities[2] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
			} else if mem.DeviceLocator == "DIMM 2" {
				capacities[3] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
			}
		case "BANK 0", "ChannelA-DIMM0", "Node0":
			capacities[0] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
		case "BANK 1", "ChannelA-DIMM1", "Node1":
			capacities[1] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
		case "BANK 2", "ChannelB-DIMM0":
			capacities[2] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
		case "BANK 3", "ChannelB-DIMM1":
			capacities[3] = fmt.Sprintf("%v GB", convertBytesToGB(mem.Capacity))
		}
		if manufacturer != mem.Manufacturer {
			manufacturer = mem.Manufacturer
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
		switch mem.SMBIOSMemoryType {
		case 26:
			memType = "DDR4"
		case 34:
			memType = "DDR5"
		}
	}

	systemInfo.Memory = Memory{
		Type:       memType,
		Speed:      int(speed),
		Name:       strings.TrimSpace(fmt.Sprintf("%v - %v", manufacturer, partNumber)),
		Capacities: capacities,
	}

	// Query WMI for CPU information
	var cpu []Win32Processor
	if err := wmi.Query("SELECT * FROM Win32_Processor", &cpu); err != nil {
		fmt.Println("Error querying Win32_Processor:", err)
		return
	}

	fmt.Println("Got CPU information")
	for _, processor := range cpu {
		fmt.Println(fmt.Sprintf("%s", strings.TrimSpace(processor.Name)))
		systemInfo.Cpu = fmt.Sprintf("%s", strings.TrimSpace(processor.Name))
	}

	// Query WMI for GPU information
	var gpu []Win32VideoController
	if err := wmi.Query("SELECT * FROM Win32_VideoController", &gpu); err != nil {
		fmt.Println("Error querying Win32_VideoController:", err)
		return
	}

	fmt.Println("Got GPU information")

	switch len(gpu) {
	case 0:
		fmt.Println("No GPU information")
		return
	case 1:
		video := gpu[0]
		systemInfo.Gpu = fmt.Sprintf("%s, %s", video.Name, video.DriverVersion)
		fmt.Println(fmt.Sprintf("%s, %s", video.Name, video.DriverVersion))
	default:
		fmt.Println(fmt.Sprintf("GPU's found: %v", len(gpu)))
		video := gpu[0]
		systemInfo.Gpu = fmt.Sprintf("%s, %s", video.Name, video.DriverVersion)
		for _, controller := range gpu[1:] {
			fmt.Println(fmt.Sprintf("Other GPU Found %s, %s", controller.Name, controller.DriverVersion))
		}
	}

	maj, _, patch := windows.RtlGetNtVersionNumbers()
	systemInfo.OperatingSystem = OperatingSystem{
		Version: fmt.Sprintf("Windows %v", strconv.Itoa(int(maj))),
		Patch:   strconv.Itoa(int(patch)),
	}

	jsonParts, _ := json.Marshal(systemInfo)
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
	SMBIOSMemoryType     uint64
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
