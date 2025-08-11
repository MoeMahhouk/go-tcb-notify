package eth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func LoadABIFromFile(path string) (abi.ABI, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("read abi file: %w", err)
	}
	txt := strings.TrimSpace(string(raw))

	// Handle either plain array or {"abi":[...]} (foundry-style) files.
	var wrap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrap); err == nil {
		if ab, ok := wrap["abi"]; ok {
			return abi.JSON(strings.NewReader(string(ab)))
		}
	}
	return abi.JSON(strings.NewReader(txt))
}
