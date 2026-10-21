package qrgen

import (
	"fmt"
	"strings"

	"rsc.io/qr"

	"fana-karaoke/internal/ports"
)

const quietZone = 4

// Generator renders QR codes as SVG.
type Generator struct{}

// New builds the generator.
func New() *Generator { return &Generator{} }

// EncodeSVG renders target as a crisp SVG QR code.
func (Generator) EncodeSVG(target string) ([]byte, error) {
	code, err := qr.Encode(target, qr.M)
	if err != nil {
		return nil, err
	}
	size := code.Size + quietZone*2
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" shape-rendering="crispEdges">`, size, size)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#fff"/>`, size, size)
	b.WriteString(`<path fill="#08050f" d="`)
	writeModules(&b, code)
	b.WriteString(`"/></svg>`)
	return []byte(b.String()), nil
}

func writeModules(b *strings.Builder, code *qr.Code) {
	for y := 0; y < code.Size; y++ {
		for x := 0; x < code.Size; x++ {
			if code.Black(x, y) {
				fmt.Fprintf(b, "M%d %dh1v1h-1z", x+quietZone, y+quietZone)
			}
		}
	}
}

var _ ports.QRGenerator = (*Generator)(nil)
