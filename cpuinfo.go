package main

import (
	"fmt"

	"github.com/earentir/cpuid"
)

func collectCPUInfo() (*CPUInfo, error) {
	// Use cpuid package to collect ALL available information
	maxFunc, maxExtFunc := cpuid.GetMaxFunctions(false, "")
	vendorID := cpuid.GetVendorID(false, "")
	vendorName := cpuid.GetVendorName(false, "")
	brandString := cpuid.GetBrandString(maxExtFunc, false, "")
	modelData := cpuid.GetModelData(false, "")
	processorInfo := cpuid.GetProcessorInfo(maxFunc, maxExtFunc, false, "")

	// Get ALL supported features with detailed information
	supportedFeatures := []string{}
	featureCategories := make(map[string][]FeatureDetail)
	categories := cpuid.GetAllFeatureCategories()
	detailedFeatures := cpuid.GetAllFeatureCategoriesDetailed()

	for _, category := range categories {
		features := cpuid.GetSupportedFeatures(category, false, "")
		supportedFeatures = append(supportedFeatures, features...)

		// Get detailed feature information
		if categoryDetails, ok := detailedFeatures[category]; ok {
			featureDetails := []FeatureDetail{}
			for _, feat := range categoryDetails {
				featureDetails = append(featureDetails, FeatureDetail{
					Name:        feat["name"],
					Description: feat["description"],
					Vendor:      feat["vendor"],
					Category:    category,
				})
			}
			featureCategories[category] = featureDetails
		}
	}

	// Get detailed cache info
	cacheInfo := []string{}
	cacheDetails := []CacheDetail{}
	caches, err := cpuid.GetCacheInfo(maxFunc, maxExtFunc, vendorID, false, "")
	if err == nil {
		for _, cache := range caches {
			// Format cache information
			cacheStr := fmt.Sprintf("L%d %s: %d KB, %d-way, %d bytes/line",
				cache.Level, cache.Type, cache.SizeKB, cache.Ways, cache.LineSizeBytes)
			cacheInfo = append(cacheInfo, cacheStr)

			// Store detailed cache info
			cacheDetails = append(cacheDetails, CacheDetail{
				Level:            cache.Level,
				Type:             cache.Type,
				SizeKB:           cache.SizeKB,
				Ways:             cache.Ways,
				LineSizeBytes:    cache.LineSizeBytes,
				TotalSets:        cache.TotalSets,
				MaxCoresSharing:  cache.MaxCoresSharing,
				SelfInitializing: cache.SelfInitializing,
				FullyAssociative: cache.FullyAssociative,
				MaxProcessorIDs:  cache.MaxProcessorIDs,
				WritePolicy:      cache.WritePolicy,
			})
		}
	}

	// Get TLB info
	tlbInfo := TLBInfo{}
	tlb, tlbErr := cpuid.GetTLBInfo(maxFunc, maxExtFunc, false, "")
	if tlbErr == nil {
		// Convert TLBLevel to TLBEntry slices - L1 has Data and Instruction, L2 has Unified
		tlbInfo.L1Data = convertTLBEntries(tlb.L1.Data)
		tlbInfo.L1Inst = convertTLBEntries(tlb.L1.Instruction)
		tlbInfo.L2Unified = convertTLBEntries(tlb.L2.Unified)
		// Also add L3 if available
		if len(tlb.L3.Unified) > 0 {
			tlbInfo.L2Unified = append(tlbInfo.L2Unified, convertTLBEntries(tlb.L3.Unified)...)
		}
	}

	// Get Hybrid info (Intel)
	hybridInfo := HybridInfo{}
	hybrid := cpuid.GetIntelHybrid(false, "")
	hybridInfo.IsHybrid = hybrid.HybridCPU
	if hybrid.HybridCPU {
		if hybrid.CoreTypeName != "" {
			hybridInfo.CoreType = hybrid.CoreTypeName
		} else if hybrid.CoreType == 0 {
			hybridInfo.CoreType = "P-core (Performance)"
		} else if hybrid.CoreType == 1 {
			hybridInfo.CoreType = "E-core (Efficient)"
		} else {
			hybridInfo.CoreType = fmt.Sprintf("Unknown (%d)", hybrid.CoreType)
		}
	}

	// Extract model information
	family := modelData.ExtendedFamily
	modelNum := modelData.ExtendedModel
	stepping := modelData.SteppingID
	cores := processorInfo.CoreCount
	threads := processorInfo.ThreadPerCore * processorInfo.CoreCount

	return &CPUInfo{
		Vendor:            vendorName,
		Brand:             brandString,
		Model:             fmt.Sprintf("Family %d, Model %d, Stepping %d", family, modelNum, stepping),
		Family:            family,
		ModelNumber:       modelNum,
		Stepping:          stepping,
		Cores:             cores,
		Threads:           threads,
		Features:          supportedFeatures,
		FeatureCategories: featureCategories,
		CacheInfo:         cacheInfo,
		CacheDetails:      cacheDetails,
		TLBInfo:           tlbInfo,
		HybridInfo:        hybridInfo,
		ProcessorInfo: ProcessorInfoDetail{
			MaxLogicalProcessors: processorInfo.MaxLogicalProcessors,
			InitialAPICID:        processorInfo.InitialAPICID,
			PhysicalAddressBits:  processorInfo.PhysicalAddressBits,
			LinearAddressBits:    processorInfo.LinearAddressBits,
			CoreCount:            processorInfo.CoreCount,
			ThreadPerCore:        processorInfo.ThreadPerCore,
		},
		ModelData: ModelDataDetail{
			SteppingID:       modelData.SteppingID,
			ModelID:          modelData.ModelID,
			FamilyID:         modelData.FamilyID,
			ProcessorType:    modelData.ProcessorType,
			ExtendedModelID:  modelData.ExtendedModelID,
			ExtendedFamilyID: modelData.ExtendedFamilyID,
			ExtendedModel:    modelData.ExtendedModel,
			ExtendedFamily:   modelData.ExtendedFamily,
		},
		MaxFunc:          maxFunc,
		MaxExtFunc:       maxExtFunc,
		PhysicalAddrBits: processorInfo.PhysicalAddressBits,
		LinearAddrBits:   processorInfo.LinearAddressBits,
	}, nil
}

func convertTLBEntries(entries []cpuid.TLBEntry) []TLBEntry {
	result := []TLBEntry{}
	for _, e := range entries {
		result = append(result, TLBEntry{
			PageSize:      e.PageSize,
			Entries:       e.Entries,
			Associativity: e.Associativity,
		})
	}
	return result
}
