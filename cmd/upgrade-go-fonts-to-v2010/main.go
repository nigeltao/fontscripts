// upgrade-go-fonts-to-v2010 upgrades the Go Fonts from version 2.008 to 2.010.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

var filenames = []string{
	"Go-Bold-Italic",
	"Go-Bold",
	"Go-Italic",
	"Go-Medium-Italic",
	"Go-Medium",
	"Go-Mono-Bold-Italic",
	"Go-Mono-Bold",
	"Go-Mono-Italic",
	"Go-Mono",
	"Go-Regular",
	"Go-Smallcaps-Italic",
	"Go-Smallcaps",
}

func main() {
	for _, filename := range filenames {
		do(filename)
	}
}

func do(filename string) {
	println(filename)

	home, _ := os.LookupEnv("HOME")
	inTTFFilename := home + "/go/src/golang.org/x/image/font/gofont/ttfs/" + filename + ".ttf"
	inTTXFilename := "/tmp/newgofont-0.ttx"
	if err := exec.Command("ttx", "-o", inTTXFilename, inTTFFilename).Run(); err != nil {
		log.Fatalf("ttx (in): %v", err)
	}

	input, err := os.ReadFile(inTTXFilename)
	if err != nil {
		log.Fatal(err)
	}
	hmtx, input := loadHmtx(input)
	glyf, input := loadGlyf(input)
	input = bumpVersionNumber(input)

	if strings.Contains(filename, "Smallcaps") {
		// Fix a copy/pasto in the hand-made Smallcaps fonts.
		swapUAcuteCircumflex(glyf)
	}

	synthesizeNewGlyphs(filename, hmtx, glyf)

	w := &bytes.Buffer{}
	skipHints := false
	inCmap := false
	inExtraNames := false
	for remaining0 := []byte(nil); len(input) > 0; input = remaining0 {
		line := input
		remaining0 = nil
		if i := bytes.IndexByte(input, '\n'); i >= 0 {
			line, remaining0 = input[:i+1], input[i+1:]
		}
		trimLine := bytes.TrimSpace(line)

		if skipHints {
			if bytes.Equal(trimLine, cvt1) ||
				bytes.Equal(trimLine, fpgm1) ||
				bytes.Equal(trimLine, prep1) {
				skipHints = false
			}
			continue
		}

		if bytes.Equal(trimLine, cvt0) ||
			bytes.Equal(trimLine, fpgm0) ||
			bytes.Equal(trimLine, prep0) {
			skipHints = true
			continue

		} else if bytes.Equal(trimLine, cmap0) {
			inCmap = true
		} else if inCmap {
			if bytes.Equal(trimLine, cmap1) {
				inCmap = false
			} else if bytes.Equal(trimLine, mapCode1fa) {
				w.Write(indent6)
				w.Write(mapCode1cdEtc)
				w.WriteByte('\n')
			} else if bytes.Equal(trimLine, mapCode384) {
				w.Write(indent6)
				w.Write(mapCode37e)
				w.WriteByte('\n')
			} else if bytes.Equal(trimLine, mapCode20a3) {
				w.Write(indent6)
				w.Write(mapCode2070Etc)
				w.WriteByte('\n')
			} else if bytes.Equal(trimLine, mapCode222b) {
				w.Write(indent6)
				w.Write(mapCode222a)
				w.WriteByte('\n')
			}

		} else if bytes.Equal(trimLine, extraNames0) {
			inExtraNames = true
		} else if inExtraNames {
			if bytes.Equal(trimLine, extraNames1) {
				inExtraNames = false
			} else if bytes.Equal(trimLine, psNameUni0218) {
				w.Write(indent6)
				w.Write(psNameUni01CDEtc)
				w.WriteByte('\n')
			} else if bytes.Equal(trimLine, psNameUni0394) {
				w.Write(indent6)
				w.Write(psNameUni037E)
				w.WriteByte('\n')
			} else if bytes.Equal(trimLine, psNameUni2105) {
				w.Write(indent6)
				w.Write(psNameUni2070Etc)
				w.WriteByte('\n')
			} else if bytes.Equal(trimLine, psNameUogonek) {
				w.Write(indent6)
				w.Write(psNameUnion)
				w.WriteByte('\n')
			}

		} else if bytes.Equal(trimLine, glyf1) {
			writeMap(w, glyf)

		} else if bytes.Equal(trimLine, glyphOrder1) {
			w.WriteString(`    <GlyphID id="999" name="uni01CD"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01CE"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01CF"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D0"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D1"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D2"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D3"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D4"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D5"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D6"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D7"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D8"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01D9"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01DA"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01DB"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni01DC"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni037E"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2070"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2074"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2075"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2076"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2077"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2078"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2079"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni207A"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni207B"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni207C"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni207D"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni207E"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2080"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2081"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2082"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2083"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2084"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2085"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2086"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2087"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2088"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2089"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni208A"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni208B"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni208C"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni208D"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni208E"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="uni2099"/>` + "\n")
			w.WriteString(`    <GlyphID id="999" name="union"/>` + "\n")

		} else if bytes.Equal(trimLine, hmtx1) {
			writeMap(w, hmtx)

		}

		w.Write(line)
	}

	out1TTXFilename := "/tmp/newgofont-1.ttx"
	out1TTFFilename := "/tmp/newgofont-1.ttf"
	if err := os.WriteFile(out1TTXFilename, w.Bytes(), 0600); err != nil {
		log.Fatal(err)
	}
	if err := exec.Command("ttx", "-o", out1TTFFilename, out1TTXFilename).Run(); err != nil {
		log.Fatalf("ttx (out1): %v", err)
	}

	// ttfreindex is github.com/nigeltao/fontscripts/cmd/ttfreindex
	out2TTFFilename := "/tmp/newgofont-2.ttf"
	if err := exec.Command("ttfreindex", "-dst", out2TTFFilename, "-src", out1TTFFilename).Run(); err != nil {
		log.Fatalf("ttfreindex: %v", err)
	}

	out3TTFFilename := "/tmp/" + filename + ".ttf"
	if err := exec.Command("ttfautohint", out2TTFFilename, out3TTFFilename).Run(); err != nil {
		log.Fatalf("ttfautohint", err)
	}

	os.Remove(inTTXFilename)
	os.Remove(out1TTXFilename)
	os.Remove(out1TTFFilename)
	os.Remove(out2TTFFilename)
}

func loadHmtx(input []byte) (hmtx map[string][]byte, remaining []byte) {
	hmtx0 := []byte("  <hmtx>\n")
	hmtx1 := []byte("  </hmtx>\n")

	hmtx = map[string][]byte{}
	if i := bytes.Index(input, hmtx0); i < 0 {
		log.Fatal("could not find hmtx0")
	} else {
		remaining = append(remaining, input[:i+len(hmtx0)]...)
		input = input[i+len(hmtx0):]
	}
	if i := bytes.Index(input, hmtx1); i < 0 {
		log.Fatal("could not find hmtx1")
	} else {
		remaining = append(remaining, input[i:]...)
		input = input[:i]
	}

	for {
		i := bytes.IndexByte(input, '\n')
		if i < 0 {
			break
		}
		line := input[:i+1]
		input = input[i+1:]
		name := parseString(line, nameQuote)
		hmtx[name] = line
	}

	return hmtx, remaining
}

func loadGlyf(input []byte) (glyf map[string][]byte, remaining []byte) {
	glyf0 := []byte("  <glyf>\n")
	glyf1 := []byte("  </glyf>\n")
	ttGlyph0 := []byte("    <TTGlyph ")
	ttGlyph1 := []byte("    </TTGlyph>\n\n")
	instructions0 := []byte("<instructions>")

	glyf = map[string][]byte{}
	if i := bytes.Index(input, glyf0); i < 0 {
		log.Fatal("could not find glyf0")
	} else {
		remaining = append(remaining, input[:i+len(glyf0)]...)
		input = input[i+len(glyf0):]
	}
	if i := bytes.Index(input, glyf1); i < 0 {
		log.Fatal("could not find glyf1")
	} else {
		remaining = append(remaining, input[i:]...)
		input = input[:i]
	}

	for {
		i := bytes.Index(input, ttGlyph0)
		j := bytes.Index(input, ttGlyph1)
		if (i < 0) || (j < 0) || (j < i) {
			break
		}
		snippet := input[i : j+len(ttGlyph1)]
		input = input[j+len(ttGlyph1):]

		name := parseString(snippet, nameQuote)

		if r := bytes.Index(snippet, instructions0); r >= 0 {
			value := []byte(nil)
			value = append(value, snippet[:r]...)
			value = append(value, "<instructions/>\n    </TTGlyph>\n\n"...)
			glyf[name] = value
		} else {
			glyf[name] = snippet
		}
	}

	return glyf, remaining
}

func bumpVersionNumber(input []byte) []byte {
	input = bytes.ReplaceAll(input,
		[]byte("Version 2.008; ttfautohint (v1.6)"),
		[]byte("Version 2.010"),
	)
	input = bytes.ReplaceAll(input,
		[]byte("fontRevision value=\"2.007"),
		[]byte("fontRevision value=\"2.010"),
	)
	input = bytes.ReplaceAll(input,
		[]byte("fontRevision value=\"2.008"),
		[]byte("fontRevision value=\"2.010"),
	)
	return input
}

func writeMap(w *bytes.Buffer, m map[string][]byte) {
	keys := []string(nil)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		w.Write(m[k])
	}
}

func parseGlyph(b []byte) (g glyph) {
	inContour := false
	for {
		line := []byte(nil)
		if i := bytes.IndexByte(b, '\n'); i < 0 {
			break
		} else {
			line = bytes.TrimSpace(b[:i])
			b = b[i+1:]
		}

		if !inContour {
			inContour = bytes.Equal(line, contour0)
			if inContour {
				g = append(g, nil)
			}
			continue
		} else if bytes.Equal(line, contour1) {
			inContour = false
			continue
		}
		c := &g[len(g)-1]
		*c = append(*c, pt{
			x:  parseInt(line, xQuote),
			y:  parseInt(line, yQuote),
			on: parseInt(line, onQuote),
		})
	}
	return g
}

func parseInt(b []byte, prefix []byte) int {
	n, _ := strconv.Atoi(parseString(b, prefix))
	return n
}

func parseString(b []byte, prefix []byte) string {
	if i := bytes.Index(b, prefix); i < 0 {
		log.Fatalf("could not find %q", prefix)
	} else {
		b = b[i+len(prefix):]
	}
	if j := bytes.IndexByte(b, '"'); j < 0 {
		log.Fatal("could not find quote")
	} else {
		b = b[:j]
	}
	return string(b)
}

func swapUAcuteCircumflex(glyf map[string][]byte) {
	g0 := parseGlyph(glyf["uacute"])
	g1 := parseGlyph(glyf["ucircumflex"])
	glyf["uacute"] = g1.render("uacute")
	glyf["ucircumflex"] = g0.render("ucircumflex")
}

func synthesizeNewGlyphs(filename string, hmtx map[string][]byte, glyf map[string][]byte) {
	italic := strings.Contains(filename, "Italic")
	mono := strings.HasPrefix(filename, "Go-Mono")

	{ // Pinyin diacritic-vowel combinations.
		caronUpper := extractDiacritic(glyf, "Scaron")
		caronLower := extractDiacritic(glyf, "scaron")
		macronUpper := extractDiacritic(glyf, "Umacron")
		macronLower := extractDiacritic(glyf, "umacron")
		acuteUpper := extractDiacritic(glyf, "Uacute")
		acuteLower := extractDiacritic(glyf, "uacute")
		graveUpper := extractDiacritic(glyf, "Ugrave")
		graveLower := extractDiacritic(glyf, "ugrave")
		synthesizePinyin(hmtx, glyf, italic, "A", caronUpper, "uni01CD")
		synthesizePinyin(hmtx, glyf, italic, "a", caronLower, "uni01CE")
		synthesizePinyin(hmtx, glyf, italic, "I", caronUpper, "uni01CF")
		synthesizePinyin(hmtx, glyf, italic, "dotlessi", caronLower, "uni01D0")
		synthesizePinyin(hmtx, glyf, italic, "O", caronUpper, "uni01D1")
		synthesizePinyin(hmtx, glyf, italic, "o", caronLower, "uni01D2")
		synthesizePinyin(hmtx, glyf, italic, "U", caronUpper, "uni01D3")
		synthesizePinyin(hmtx, glyf, italic, "u", caronLower, "uni01D4")
		synthesizePinyin(hmtx, glyf, italic, "Udieresis", macronUpper, "uni01D5")
		synthesizePinyin(hmtx, glyf, italic, "udieresis", macronLower, "uni01D6")
		synthesizePinyin(hmtx, glyf, italic, "Udieresis", acuteUpper, "uni01D7")
		synthesizePinyin(hmtx, glyf, italic, "udieresis", acuteLower, "uni01D8")
		synthesizePinyin(hmtx, glyf, italic, "Udieresis", caronUpper, "uni01D9")
		synthesizePinyin(hmtx, glyf, italic, "udieresis", caronLower, "uni01DA")
		synthesizePinyin(hmtx, glyf, italic, "Udieresis", graveUpper, "uni01DB")
		synthesizePinyin(hmtx, glyf, italic, "udieresis", graveLower, "uni01DC")
	}

	{ // U+037E GREEK QUESTION MARK
		hmtx["uni037E"] = bytes.Replace(hmtx["semicolon"], []byte("semicolon"), []byte("uni037E"), 1)
		glyf["uni037E"] = parseGlyph(glyf["semicolon"]).render("uni037E")
	}

	{ // U+222A UNION
		w := parseInt(hmtx["intersection"], widthQuote)
		g := parseGlyph(glyf["intersection"])
		xMin, yMin, _, yMax := g.bounds()

		centerLo := 0
		centerHi := 0
		for _, c := range g {
			for _, p := range c {
				if p.y == 0 {
					centerLo += p.x
				} else if p.y == 1480 {
					centerHi += p.x
				}
			}
		}
		italicCorrection := (centerHi / 3) - (centerLo / 4)

		// Subtract italicCorrection.
		if italicCorrection != 0 {
			for _, c := range g {
				for j := range c {
					c[j].x -= ((c[j].y - yMin) * italicCorrection) / (yMax - yMin)
				}
			}
		}
		// Flip vertical.
		for _, c := range g {
			for j := range c {
				c[j].y = yMax + yMin - c[j].y
			}
		}
		// Add italicCorrection.
		if italicCorrection != 0 {
			for _, c := range g {
				for j := range c {
					c[j].x += ((c[j].y - yMin) * italicCorrection) / (yMax - yMin)
				}
			}
		}

		hmtx["union"] = []byte(fmt.Sprintf(`    <mtx name="union" width="%d" lsb="%d"/>`+"\n", w, xMin))
		glyf["union"] = g.render("union")
	}

	if strings.HasPrefix(filename, "Go-Medium") {
		patchKnobblyL(glyf, "l")
		patchKnobblyL(glyf, "lacute")
		patchKnobblyL(glyf, "lcaron")
		patchKnobblyL(glyf, "ldot")
		patchKnobblyL(glyf, "lslash")
		patchKnobblyL(glyf, "uni013C")
	}

	{
		width := parseInt(hmtx["uni207F"], widthQuote) // U+207F SUPERSCRIPT LATIN SMALL LETTER N.
		_, yAdjust, _, _ := parseGlyph(glyf["uni207F"]).bounds()

		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2070", "uni2080", "zero")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni00B9", "uni2081", "one")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni00B2", "uni2082", "two")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni00B3", "uni2083", "three")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2074", "uni2084", "four")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2075", "uni2085", "five")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2076", "uni2086", "six")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2077", "uni2087", "seven")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2078", "uni2088", "eight")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni2079", "uni2089", "nine")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni207A", "uni208A", "plus")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni207B", "uni208B", "minus")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni207C", "uni208C", "equal")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni207D", "uni208D", "parenleft")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni207E", "uni208E", "parenright")
		synthesizeSuperscriptSubscript(hmtx, glyf, mono, italic, width, yAdjust, "uni207F", "uni2099", "n")
	}
}

func extractDiacritic(glyf map[string][]byte, name string) contour {
	for _, c := range parseGlyph(glyf[name]) {
		if len(c) > 8 {
			continue
		}
		c = c.clone()
		_, yMin, _, _ := c.bounds()
		center := 0
		for _, p := range c {
			if p.y == yMin {
				center += p.x
			}
		}
		center /= 2
		c.nudge(-center, 0)
		return c
	}
	return nil
}

func patchKnobblyL(glyf map[string][]byte, name string) {
	g := parseGlyph(glyf[name])
	for i, c := range g {
		if len(c) < 1 {
			continue
		}

		switch c[0] {
		// Go-Medium.
		case pt{x: 391, y: 355, on: 1}, pt{x: 411, y: 355, on: 1}:
			xDelta := c[0].x - 391
			keep := -1
			for j, p := range c {
				if p == (pt{x: 557 + xDelta, y: -9, on: 1}) {
					keep = j
					break
				}
			}
			if keep < 0 {
				continue
			}
			g[i] = nil
			g[i] = append(g[i],
				pt{x: 391 + xDelta, y: 355, on: 1},
				pt{x: 391 + xDelta, y: 223, on: 0},
				pt{x: 462 + xDelta, y: 152, on: 0},
				pt{x: 544 + xDelta, y: 152, on: 1},
				pt{x: 557 + xDelta, y: 152, on: 1},
			)
			g[i] = append(g[i], c[keep:]...)

		// Go-Medium-Italic. Gradient is dy/dx = 5/1.
		case pt{x: 461, y: 355, on: 1}, pt{x: 481, y: 355, on: 1}:
			xDelta := c[0].x - 461
			keep := -1
			for j, p := range c {
				if p == (pt{x: 555 + xDelta, y: -9, on: 1}) {
					keep = j
					break
				}
			}
			if keep < 0 {
				continue
			}
			g[i] = nil
			g[i] = append(g[i],
				pt{x: 461 + xDelta, y: 355, on: 1},
				pt{x: 435 + xDelta, y: 223, on: 0},
				pt{x: 492 + xDelta, y: 152, on: 0},
				pt{x: 574 + xDelta, y: 152, on: 1},
				pt{x: 587 + xDelta, y: 152, on: 1},
			)
			g[i] = append(g[i], c[keep:]...)
		}
	}
	glyf[name] = g.render(name)
}

func synthesizePinyin(hmtx map[string][]byte, glyf map[string][]byte, italic bool, base string, diacritic contour, dst string) {
	deltaY := 0
	if (base == "Udieresis") || (base == "udieresis") {
		deltaY = 356
	}
	w := parseInt(hmtx[base], widthQuote)
	g := parseGlyph(glyf[base])
	xMin0, _, xMax0, _ := g.italicCorrectedBounds(italic)
	d := diacritic.clone()
	d.nudge((xMin0+xMax0)/2, deltaY)
	if italic {
		// Italic gradient is dy/dx = 5/1.
		_, yMin1, _, _ := d.bounds()
		d.nudge(yMin1/5, 0)
	}
	g = append(g, d)
	xMin2, _, _, _ := g.bounds()
	hmtx[dst] = []byte(fmt.Sprintf(`    <mtx name="%s" width="%d" lsb="%d"/>`+"\n", dst, w, xMin2))
	glyf[dst] = g.render(dst)
}

func synthesizeSuperscriptSubscript(
	hmtx map[string][]byte,
	glyf map[string][]byte,
	mono bool,
	italic bool,
	width int,
	yAdjust int,
	sup string,
	sub string,
	level string) {

	xAdjust := 0
	yAdjustSup := +yAdjust
	yAdjustSub := -yAdjust / 3
	if mono {
		xAdjust = (width * 1) / 8
	} else {
		yAdjustSup = (yAdjust * 3) / 4
		if strings.HasPrefix(level, "paren") {
			xAdjust = (width * 1) / 8
		} else {
			width = (width * 5) / 4
		}
	}

	xAdjustSup := xAdjust
	xAdjustSub := xAdjust

	if !mono && italic {
		xAdjustSup += (width * 3) / 16
		xAdjustSub -= (width * 1) / 16
	}

	{
		g := parseGlyph(glyf[level])

		// Scale and translate.
		for _, c := range g {
			for j, p := range c {
				c[j].x = ((p.x * 3) / 4) + xAdjustSup
				c[j].y = ((p.y * 3) / 5) + yAdjustSup
			}
		}

		xMin, _, _, _ := g.bounds()
		hmtx[sup] = []byte(fmt.Sprintf(`    <mtx name="%s" width="%d" lsb="%d"/>`+"\n", sup, width, xMin))
		glyf[sup] = g.render(sup)
	}

	{
		g := parseGlyph(glyf[level])

		// Scale and translate.
		for _, c := range g {
			for j, p := range c {
				c[j].x = ((p.x * 3) / 4) + xAdjustSub
				c[j].y = ((p.y * 3) / 5) + yAdjustSub
			}
		}

		xMin, _, _, _ := g.bounds()
		hmtx[sub] = []byte(fmt.Sprintf(`    <mtx name="%s" width="%d" lsb="%d"/>`+"\n", sub, width, xMin))
		glyf[sub] = g.render(sub)
	}
}

func cloneBytes(b []byte) []byte {
	return append([]byte(nil), b...)
}

type pt struct {
	x, y, on int
}

type contour []pt

func (c contour) clone() contour {
	return append(contour(nil), c...)
}

func (c contour) bounds() (xMin int, yMin int, xMax int, yMax int) {
	first := true
	for _, p := range c {
		if first {
			xMin = p.x
			yMin = p.y
			xMax = p.x
			yMax = p.y
			first = false
			continue
		}

		if xMin > p.x {
			xMin = p.x
		}
		if yMin > p.y {
			yMin = p.y
		}
		if xMax < p.x {
			xMax = p.x
		}
		if yMax < p.y {
			yMax = p.y
		}
	}
	return xMin, yMin, xMax, yMax
}

func (c contour) nudge(dx int, dy int) {
	for j := range c {
		c[j].x += dx
		c[j].y += dy
	}
}

type glyph []contour

func (g glyph) bounds() (xMin int, yMin int, xMax int, yMax int) {
	first := true
	for _, c := range g {
		for _, p := range c {
			if first {
				xMin = p.x
				yMin = p.y
				xMax = p.x
				yMax = p.y
				first = false
				continue
			}

			if xMin > p.x {
				xMin = p.x
			}
			if yMin > p.y {
				yMin = p.y
			}
			if xMax < p.x {
				xMax = p.x
			}
			if yMax < p.y {
				yMax = p.y
			}
		}
	}
	return xMin, yMin, xMax, yMax
}

func (g glyph) italicCorrectedBounds(correct bool) (xMin int, yMin int, xMax int, yMax int) {
	if !correct {
		return g.bounds()
	}

	// Italic gradient is dy/dx = 5/1.
	first := true
	for _, c := range g {
		for _, p := range c {
			if first {
				xMin = p.x - (p.y / 5)
				yMin = p.y
				xMax = p.x - (p.y / 5)
				yMax = p.y
				first = false
				continue
			}

			if xMin > p.x-(p.y/5) {
				xMin = p.x - (p.y / 5)
			}
			if yMin > p.y {
				yMin = p.y
			}
			if xMax < p.x-(p.y/5) {
				xMax = p.x - (p.y / 5)
			}
			if yMax < p.y {
				yMax = p.y
			}
		}
	}
	return xMin, yMin, xMax, yMax
}

func (g glyph) render(name string) []byte {
	b := &bytes.Buffer{}
	xMin, yMin, xMax, yMax := g.bounds()
	fmt.Fprintf(b, `    <TTGlyph name="%s" xMin="%d" yMin="%d" xMax="%d" yMax="%d">`+"\n",
		name, xMin, yMin, xMax, yMax)

	for _, c := range g {
		b.WriteString("      <contour>\n")
		for _, p := range c {
			fmt.Fprintf(b, `        <pt x="%d" y="%d" on="%d"/>`+"\n", p.x, p.y, p.on)
		}
		b.WriteString("      </contour>\n")
	}

	b.WriteString("      <instructions/>\n    </TTGlyph>\n\n")
	return b.Bytes()
}

var (
	cmap0       = []byte("<cmap>")
	cmap1       = []byte("</cmap>")
	contour0    = []byte("<contour>")
	contour1    = []byte("</contour>")
	cvt0        = []byte("<cvt>")
	cvt1        = []byte("</cvt>")
	fpgm0       = []byte("<fpgm>")
	fpgm1       = []byte("</fpgm>")
	extraNames0 = []byte("<extraNames>")
	extraNames1 = []byte("</extraNames>")
	glyf1       = []byte("</glyf>")
	glyphOrder1 = []byte("</GlyphOrder>")
	hmtx1       = []byte("</hmtx>")
	prep0       = []byte("<prep>")
	prep1       = []byte("</prep>")

	nameQuote  = []byte(`name="`)
	onQuote    = []byte(`on="`)
	widthQuote = []byte(`width="`)
	xQuote     = []byte(`x="`)
	yQuote     = []byte(`y="`)

	mapCode1cdEtc = []byte(`` +
		`<map code="0x1cd" name="uni01CD"/>` + "\n      " +
		`<map code="0x1ce" name="uni01CE"/>` + "\n      " +
		`<map code="0x1cf" name="uni01CF"/>` + "\n      " +
		`<map code="0x1d0" name="uni01D0"/>` + "\n      " +
		`<map code="0x1d1" name="uni01D1"/>` + "\n      " +
		`<map code="0x1d2" name="uni01D2"/>` + "\n      " +
		`<map code="0x1d3" name="uni01D3"/>` + "\n      " +
		`<map code="0x1d4" name="uni01D4"/>` + "\n      " +
		`<map code="0x1d5" name="uni01D5"/>` + "\n      " +
		`<map code="0x1d6" name="uni01D6"/>` + "\n      " +
		`<map code="0x1d7" name="uni01D7"/>` + "\n      " +
		`<map code="0x1d8" name="uni01D8"/>` + "\n      " +
		`<map code="0x1d9" name="uni01D9"/>` + "\n      " +
		`<map code="0x1da" name="uni01DA"/>` + "\n      " +
		`<map code="0x1db" name="uni01DB"/>` + "\n      " +
		`<map code="0x1dc" name="uni01DC"/>`)
	mapCode1fa     = []byte(`<map code="0x1fa" name="Aringacute"/><!-- LATIN CAPITAL LETTER A WITH RING ABOVE AND ACUTE -->`)
	mapCode37e     = []byte(`<map code="0x37e" name="uni037E"/>`)
	mapCode384     = []byte(`<map code="0x384" name="tonos"/><!-- GREEK TONOS -->`)
	mapCode2070Etc = []byte(`` +
		`<map code="0x2070" name="uni2070"/>` + "\n      " +
		`<map code="0x2074" name="uni2074"/>` + "\n      " +
		`<map code="0x2075" name="uni2075"/>` + "\n      " +
		`<map code="0x2076" name="uni2076"/>` + "\n      " +
		`<map code="0x2077" name="uni2077"/>` + "\n      " +
		`<map code="0x2078" name="uni2078"/>` + "\n      " +
		`<map code="0x2079" name="uni2079"/>` + "\n      " +
		`<map code="0x207A" name="uni207A"/>` + "\n      " +
		`<map code="0x207B" name="uni207B"/>` + "\n      " +
		`<map code="0x207C" name="uni207C"/>` + "\n      " +
		`<map code="0x207D" name="uni207D"/>` + "\n      " +
		`<map code="0x207E" name="uni207E"/>` + "\n      " +
		`<map code="0x2080" name="uni2080"/>` + "\n      " +
		`<map code="0x2081" name="uni2081"/>` + "\n      " +
		`<map code="0x2082" name="uni2082"/>` + "\n      " +
		`<map code="0x2083" name="uni2083"/>` + "\n      " +
		`<map code="0x2084" name="uni2084"/>` + "\n      " +
		`<map code="0x2085" name="uni2085"/>` + "\n      " +
		`<map code="0x2086" name="uni2086"/>` + "\n      " +
		`<map code="0x2087" name="uni2087"/>` + "\n      " +
		`<map code="0x2088" name="uni2088"/>` + "\n      " +
		`<map code="0x2089" name="uni2089"/>` + "\n      " +
		`<map code="0x208A" name="uni208A"/>` + "\n      " +
		`<map code="0x208B" name="uni208B"/>` + "\n      " +
		`<map code="0x208C" name="uni208C"/>` + "\n      " +
		`<map code="0x208D" name="uni208D"/>` + "\n      " +
		`<map code="0x208E" name="uni208E"/>` + "\n      " +
		`<map code="0x2099" name="uni2099"/>`)
	mapCode20a3 = []byte(`<map code="0x20a3" name="franc"/><!-- FRENCH FRANC SIGN -->`)
	mapCode222a = []byte(`<map code="0x222a" name="union"/>`)
	mapCode222b = []byte(`<map code="0x222b" name="integral"/><!-- INTEGRAL -->`)

	psNameUni01CDEtc = []byte(`` +
		`<psName name="uni01CD"/>` + "\n      " +
		`<psName name="uni01CE"/>` + "\n      " +
		`<psName name="uni01CF"/>` + "\n      " +
		`<psName name="uni01D0"/>` + "\n      " +
		`<psName name="uni01D1"/>` + "\n      " +
		`<psName name="uni01D2"/>` + "\n      " +
		`<psName name="uni01D3"/>` + "\n      " +
		`<psName name="uni01D4"/>` + "\n      " +
		`<psName name="uni01D5"/>` + "\n      " +
		`<psName name="uni01D6"/>` + "\n      " +
		`<psName name="uni01D7"/>` + "\n      " +
		`<psName name="uni01D8"/>` + "\n      " +
		`<psName name="uni01D9"/>` + "\n      " +
		`<psName name="uni01DA"/>` + "\n      " +
		`<psName name="uni01DB"/>` + "\n      " +
		`<psName name="uni01DC"/>`)
	psNameUni0218    = []byte(`<psName name="uni0218"/>`)
	psNameUni037E    = []byte(`<psName name="uni037E"/>`)
	psNameUni0394    = []byte(`<psName name="uni0394"/>`)
	psNameUni2070Etc = []byte(`` +
		`<psName name="uni2070"/>` + "\n      " +
		`<psName name="uni2074"/>` + "\n      " +
		`<psName name="uni2075"/>` + "\n      " +
		`<psName name="uni2076"/>` + "\n      " +
		`<psName name="uni2077"/>` + "\n      " +
		`<psName name="uni2078"/>` + "\n      " +
		`<psName name="uni2079"/>` + "\n      " +
		`<psName name="uni207A"/>` + "\n      " +
		`<psName name="uni207B"/>` + "\n      " +
		`<psName name="uni207C"/>` + "\n      " +
		`<psName name="uni207D"/>` + "\n      " +
		`<psName name="uni207E"/>` + "\n      " +
		`<psName name="uni2080"/>` + "\n      " +
		`<psName name="uni2081"/>` + "\n      " +
		`<psName name="uni2082"/>` + "\n      " +
		`<psName name="uni2083"/>` + "\n      " +
		`<psName name="uni2084"/>` + "\n      " +
		`<psName name="uni2085"/>` + "\n      " +
		`<psName name="uni2086"/>` + "\n      " +
		`<psName name="uni2087"/>` + "\n      " +
		`<psName name="uni2088"/>` + "\n      " +
		`<psName name="uni2089"/>` + "\n      " +
		`<psName name="uni208A"/>` + "\n      " +
		`<psName name="uni208B"/>` + "\n      " +
		`<psName name="uni208C"/>` + "\n      " +
		`<psName name="uni208D"/>` + "\n      " +
		`<psName name="uni208E"/>` + "\n      " +
		`<psName name="uni2099"/>`)
	psNameUni2105 = []byte(`<psName name="uni2105"/>`)
	psNameUnion   = []byte(`<psName name="union"/>`)
	psNameUogonek = []byte(`<psName name="uogonek"/>`)

	indent6 = []byte("      ")
)
