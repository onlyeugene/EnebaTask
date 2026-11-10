package products

import (
	"fmt"
	"hash/fnv"
	"html"
	"strings"
)

// Cover sizes match the product card image area in notes.md (222 x 250).
const (
	coverW = 222
	coverH = 250
)

// CoverSVG draws a cover from a short title ("Starfall Odyssey", not the full listing name)
// and a two-color palette.
// The same input always gives the same SVG: the shapes come from a hash of the title,
// not from a random source, so re-seeding never changes how a product looks.
func CoverSVG(title string, palette [2]string) string {
	h := fnv.New32a()
	h.Write([]byte(title))
	seed := h.Sum32()

	// next returns a number in [0, n) and moves the seed on. A tiny LCG is enough for shape placement.
	next := func(n uint32) uint32 {
		seed = seed*1664525 + 1013904223
		return (seed >> 8) % n
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, coverW, coverH, coverW, coverH)
	fmt.Fprintf(&b, `<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">`+
		`<stop offset="0" stop-color="%s"/><stop offset="1" stop-color="#1f0a4d"/></linearGradient></defs>`, palette[0])
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="url(#g)"/>`, coverW, coverH)

	// A few large shapes in the accent color give each cover its own look.
	for i := 0; i < 4; i++ {
		x, y := next(coverW), next(coverH*2/3)
		size := 30 + next(90)
		opacity := 0.15 + float64(next(40))/100
		if next(2) == 0 {
			fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="%s" fill-opacity="%.2f"/>`, x, y, size/2, palette[1], opacity)
		} else {
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" fill-opacity="%.2f" transform="rotate(45 %d %d)"/>`,
				x, y, size, size, palette[1], opacity, x+size/2, y+size/2)
		}
	}

	// Title at the bottom, wrapped into short lines. Escaped because it ends up inside XML.
	lines := wrap(title, 14)
	y := coverH - 20 - (len(lines)-1)*26
	for _, line := range lines {
		fmt.Fprintf(&b, `<text x="16" y="%d" font-family="Arial, Helvetica, sans-serif" font-size="22" font-weight="800" fill="#ffffff">%s</text>`,
			y, html.EscapeString(line))
		y += 26
	}
	b.WriteString(`</svg>`)
	return b.String()
}

// wrap splits text into lines of at most max characters, breaking only between words.
// A single word longer than max gets its own line.
func wrap(text string, max int) []string {
	var lines []string
	var cur string
	for _, w := range strings.Fields(text) {
		switch {
		case cur == "":
			cur = w
		case len(cur)+1+len(w) <= max:
			cur += " " + w
		default:
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
