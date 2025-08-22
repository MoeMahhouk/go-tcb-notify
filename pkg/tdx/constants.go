package tdx

import "fmt"

// TDX Component Names - indices 0-15 are SGX, 16-31 are TDX
var ComponentNames = struct {
	SGX []string
	TDX []string
}{
	SGX: []string{
		"SGX.CPUSVN[0]", "SGX.CPUSVN[1]", "SGX.CPUSVN[2]", "SGX.CPUSVN[3]",
		"SGX.CPUSVN[4]", "SGX.CPUSVN[5]", "SGX.CPUSVN[6]", "SGX.CPUSVN[7]",
		"SGX.CPUSVN[8]", "SGX.CPUSVN[9]", "SGX.CPUSVN[10]", "SGX.CPUSVN[11]",
		"SGX.CPUSVN[12]", "SGX.CPUSVN[13]", "SGX.CPUSVN[14]", "SGX.CPUSVN[15]",
	},
	TDX: []string{
		"TDX.SEAM_Loader", // Position 0 (index 16 overall)
		"TDX.SEAM_Module", // Position 1 (index 17 overall)
		"TDX.P_SEAM",      // Position 2 (index 18 overall)
		"TDX.TDX_Module",  // Position 3 (index 19 overall)
		"TDX.Component4",  // Position 4
		"TDX.Component5",  // Position 5
		"TDX.Component6",  // Position 6
		"TDX.Component7",  // Position 7
		"TDX.Component8",  // Position 8
		"TDX.Component9",  // Position 9
		"TDX.Component10", // Position 10
		"TDX.Component11", // Position 11
		"TDX.Component12", // Position 12
		"TDX.Component13", // Position 13
		"TDX.Component14", // Position 14
		"TDX.Component15", // Position 15
	},
}

// GetComponentName returns the name for a component at the given index
func GetComponentName(index int) string {
	if index < 16 {
		if index < len(ComponentNames.SGX) {
			return ComponentNames.SGX[index]
		}
		return fmt.Sprintf("SGX.Unknown[%d]", index)
	}

	tdxIndex := index - 16
	if tdxIndex < len(ComponentNames.TDX) {
		return ComponentNames.TDX[tdxIndex]
	}
	return fmt.Sprintf("TDX.Unknown[%d]", tdxIndex)
}

// GetComponentType returns "SGX" or "TDX" based on component index
func GetComponentType(index int) string {
	if index < 16 {
		return "SGX"
	}
	return "TDX"
}
