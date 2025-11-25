package input

import (
	"fmt"

	"github.com/vmware/govmomi/vim25/types"
)

type KeyDef struct {
	Code     int32
	Shift    bool
}

var keyMap = map[rune]KeyDef{
	'a': {0x04, false}, 'A': {0x04, true},
	'b': {0x05, false}, 'B': {0x05, true},
	'c': {0x06, false}, 'C': {0x06, true},
	'd': {0x07, false}, 'D': {0x07, true},
	'e': {0x08, false}, 'E': {0x08, true},
	'f': {0x09, false}, 'F': {0x09, true},
	'g': {0x0A, false}, 'G': {0x0A, true},
	'h': {0x0B, false}, 'H': {0x0B, true},
	'i': {0x0C, false}, 'I': {0x0C, true},
	'j': {0x0D, false}, 'J': {0x0D, true},
	'k': {0x0E, false}, 'K': {0x0E, true},
	'l': {0x0F, false}, 'L': {0x0F, true},
	'm': {0x10, false}, 'M': {0x10, true},
	'n': {0x11, false}, 'N': {0x11, true},
	'o': {0x12, false}, 'O': {0x12, true},
	'p': {0x13, false}, 'P': {0x13, true},
	'q': {0x14, false}, 'Q': {0x14, true},
	'r': {0x15, false}, 'R': {0x15, true},
	's': {0x16, false}, 'S': {0x16, true},
	't': {0x17, false}, 'T': {0x17, true},
	'u': {0x18, false}, 'U': {0x18, true},
	'v': {0x19, false}, 'V': {0x19, true},
	'w': {0x1A, false}, 'W': {0x1A, true},
	'x': {0x1B, false}, 'X': {0x1B, true},
	'y': {0x1C, false}, 'Y': {0x1C, true},
	'z': {0x1D, false}, 'Z': {0x1D, true},
	' ': {0x2C, false},
	'\n': {0x28, false},
}

func StringToUsbScanCodes(s string) ([]types.UsbScanCodeSpecKeyEvent, error) {
	var codes []types.UsbScanCodeSpecKeyEvent
	
	for _, char := range s {
		def, ok := keyMap[char]
		if !ok {
			fmt.Printf("Warning: Skipping unsupported character '%%c'\n", char)
			continue
		}

		event := types.UsbScanCodeSpecKeyEvent{
			UsbHidCode: def.Code,
		}
		if def.Shift {
			trueVal := true
			event.Modifiers = &types.UsbScanCodeSpecModifierType{
				LeftShift: &trueVal,
			}
		}
		codes = append(codes, event)
	}
	return codes, nil
}