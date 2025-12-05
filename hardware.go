package main

import (
	"fmt"
	"strings"
)

type HardwareInfo struct {
	CPU CPUInfo
}

type CPUInfo struct {
	Vendor            string
	Brand             string
	Model             string
	Family            uint32
	ModelNumber       uint32
	Stepping          uint32
	Cores             uint32
	Threads           uint32
	Features          []string
	FeatureCategories map[string][]FeatureDetail
	CacheInfo         []string
	CacheDetails      []CacheDetail
	TLBInfo           TLBInfo
	HybridInfo        HybridInfo
	ProcessorInfo     ProcessorInfoDetail
	ModelData         ModelDataDetail
	MaxFunc           uint32
	MaxExtFunc        uint32
	PhysicalAddrBits  uint32
	LinearAddrBits    uint32
}

type FeatureDetail struct {
	Name        string
	Description string
	Vendor      string
	Category    string
}

type CacheDetail struct {
	Level            uint32
	Type             string
	SizeKB           uint32
	Ways             uint32
	LineSizeBytes    uint32
	TotalSets        uint32
	MaxCoresSharing  uint32
	SelfInitializing bool
	FullyAssociative bool
	MaxProcessorIDs  uint32
	WritePolicy      string
}

type TLBInfo struct {
	L1Data    []TLBEntry
	L1Inst    []TLBEntry
	L2Unified []TLBEntry
}

type TLBEntry struct {
	PageSize      string
	Entries       int
	Associativity string
}

type HybridInfo struct {
	IsHybrid bool
	CoreType string
}

type ProcessorInfoDetail struct {
	MaxLogicalProcessors uint32
	InitialAPICID        uint32
	PhysicalAddressBits  uint32
	LinearAddressBits    uint32
	CoreCount            uint32
	ThreadPerCore        uint32
}

type ModelDataDetail struct {
	SteppingID       uint32
	ModelID          uint32
	FamilyID         uint32
	ProcessorType    uint32
	ExtendedModelID  uint32
	ExtendedFamilyID uint32
	ExtendedModel    uint32
	ExtendedFamily   uint32
}

func CollectHardwareInfo() (*HardwareInfo, error) {
	info := &HardwareInfo{}

	// Collect CPU info
	cpuInfo, err := collectCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU info: %w", err)
	}
	info.CPU = *cpuInfo

	return info, nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " " + word
			} else {
				currentLine = word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
