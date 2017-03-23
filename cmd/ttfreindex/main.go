// ttfreindex reads a TTF font file, sorts the glyphs so that the glyph order
// matches Unicode code point order, and writes out a re-indexed TTF.
package main

// TODO: don't assume that the source TTF is restricted to the BMP (Basic
// Multi-lingual Plane).

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"golang.org/x/image/font/sfnt"
)

var (
	srcFlag = flag.String("src", "", "source TTF filename")
	dstFlag = flag.String("dst", "", "destination TTF filename")
)

func main() {
	flag.Parse()
	if *srcFlag == "" || *dstFlag == "" {
		fmt.Fprintf(os.Stderr, "usage: %s -src filename1.ttf -dst filename2.ttf\n", os.Args[0])
		os.Exit(1)
	}
	dst, src := *dstFlag, *srcFlag

	srcData, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}
	srcFont, err := sfnt.Parse(srcData)
	if err != nil {
		log.Fatalf("Parse: %v", err)
	}

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("ttfreindex-%d.ttx", os.Getpid()))
	defer os.Remove(tmp)

	if err := exec.Command("ttx", "-o", tmp, src).Run(); err != nil {
		log.Fatalf("ttx (src): %v", err)
	}

	if err := ioutil.WriteFile(tmp, rewrite(tmp, srcFont), 0666); err != nil {
		log.Fatalf("WriteFile: %v", err)
	}

	if err := exec.Command("ttx", "-o", dst, tmp).Run(); err != nil {
		log.Fatalf("ttx (dst): %v", err)
	}
	fmt.Printf("Wrote %s\n", dst)
}

const notSeen rune = 0x7fffffff

type entry struct {
	oldID   int
	oldName string
	newName string
	r       rune
}

type byR []entry

func (b byR) Len() int      { return len(b) }
func (b byR) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byR) Less(i, j int) bool {
	x, y := b[i], b[j]
	if x.r != y.r {
		return x.r < y.r
	}
	return x.oldName < y.oldName
}

var (
	glyphIDRegExp = regexp.MustCompile(`^    <GlyphID id="[^"]*" name="([^"]*)"/>$`)
	nameRegExp    = regexp.MustCompile(`^( *<.*name=")([^"]*)(".*)$`)
)

func rewrite(tmpFilename string, f *sfnt.Font) []byte {
	data, err := ioutil.ReadFile(tmpFilename)
	if err != nil {
		log.Fatal(err)
	}
	lines := bytes.Split(data, []byte("\n"))

	entries := []entry(nil)
	sentinel := []byte("  </GlyphOrder>")
	for _, l := range lines {
		if x := glyphIDRegExp.FindSubmatch(l); x != nil {
			name := string(x[1])
			entries = append(entries, entry{
				oldID:   len(entries),
				oldName: name,
				newName: name,
				r:       notSeen,
			})
		}
		if bytes.Equal(l, sentinel) {
			break
		}
	}

	var buf sfnt.Buffer
	renames := map[string]string{}
	for r := rune(0); r < 0xffff; r++ {
		x, err := f.GlyphIndex(&buf, r)
		if err != nil {
			log.Fatal(err)
		}
		if x != 0 && entries[x].r == notSeen {
			entries[x].r = r

			// Don't rename the Private Use Area.
			if '\uE000' <= r && r <= '\uF8FF' {
				continue
			}

			oldName := entries[x].oldName
			newName, ok := aglfn[r]
			if !ok {
				newName = fmt.Sprintf("uni%04X", r)
			}
			entries[x].newName = newName
			if oldName != newName {
				renames[oldName] = newName
			}
		}
	}

	extraNames := []string(nil)
	for _, e := range entries {
		if !builtIns[e.newName] {
			extraNames = append(extraNames, e.newName)
		}
	}
	sort.Strings(extraNames)

	// The [1:] is because the first glyph must be .notdef.
	sort.Sort(byR(entries[1:]))

	// For debugging.
	if false {
		for i, e := range entries {
			fmt.Printf("nID=%-3d  oID=%-3d  r=%08x  on=%-20s nn=%-20s\n",
				i, e.oldID, e.r, e.oldName, e.newName)
		}
	}

	glyphOrderBytes := []byte("  <GlyphOrder>")
	extraNamesBytes := []byte("    <extraNames>")

	out := new(bytes.Buffer)
	for len(lines) > 0 {
		line := lines[0]
		lines = lines[1:]

		if x := nameRegExp.FindSubmatch(line); x != nil {
			oldName := string(x[2])
			if newName, ok := renames[oldName]; ok {
				line = nil
				line = append(line, x[1]...)
				line = append(line, newName...)
				line = append(line, x[3]...)
			}
		}

		out.Write(line)
		out.WriteByte('\n')

		if bytes.Equal(line, glyphOrderBytes) {
			lines = writeGlyphOrder(out, lines, entries)
			continue
		}
		if bytes.Equal(line, extraNamesBytes) {
			lines = writeExtraNames(out, lines, extraNames)
			continue
		}
	}
	return out.Bytes()
}

func writeGlyphOrder(out *bytes.Buffer, lines [][]byte, entries []entry) (newLines [][]byte) {
	for i, e := range entries {
		fmt.Fprintf(out, "<GlyphID id=\"%d\" name=%q/>\n", i, e.newName)
	}
	sentinel := []byte("  </GlyphOrder>")
	for ; len(lines) > 0 && !bytes.Equal(lines[0], sentinel); lines = lines[1:] {
	}
	return lines
}

func writeExtraNames(out *bytes.Buffer, lines [][]byte, extraNames []string) (newLines [][]byte) {
	for _, n := range extraNames {
		fmt.Fprintf(out, "<psName name=%q/>\n", n)
	}
	sentinel := []byte("    </extraNames>")
	for ; len(lines) > 0 && !bytes.Equal(lines[0], sentinel); lines = lines[1:] {
	}
	return lines
}

// builtIns come from
// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6post.html
var builtIns = map[string]bool{
	".notdef":          true,
	".null":            true,
	"nonmarkingreturn": true,
	"space":            true,
	"exclam":           true,
	"quotedbl":         true,
	"numbersign":       true,
	"dollar":           true,
	"percent":          true,
	"ampersand":        true,
	"quotesingle":      true,
	"parenleft":        true,
	"parenright":       true,
	"asterisk":         true,
	"plus":             true,
	"comma":            true,
	"hyphen":           true,
	"period":           true,
	"slash":            true,
	"zero":             true,
	"one":              true,
	"two":              true,
	"three":            true,
	"four":             true,
	"five":             true,
	"six":              true,
	"seven":            true,
	"eight":            true,
	"nine":             true,
	"colon":            true,
	"semicolon":        true,
	"less":             true,
	"equal":            true,
	"greater":          true,
	"question":         true,
	"at":               true,
	"A":                true,
	"B":                true,
	"C":                true,
	"D":                true,
	"E":                true,
	"F":                true,
	"G":                true,
	"H":                true,
	"I":                true,
	"J":                true,
	"K":                true,
	"L":                true,
	"M":                true,
	"N":                true,
	"O":                true,
	"P":                true,
	"Q":                true,
	"R":                true,
	"S":                true,
	"T":                true,
	"U":                true,
	"V":                true,
	"W":                true,
	"X":                true,
	"Y":                true,
	"Z":                true,
	"bracketleft":      true,
	"backslash":        true,
	"bracketright":     true,
	"asciicircum":      true,
	"underscore":       true,
	"grave":            true,
	"a":                true,
	"b":                true,
	"c":                true,
	"d":                true,
	"e":                true,
	"f":                true,
	"g":                true,
	"h":                true,
	"i":                true,
	"j":                true,
	"k":                true,
	"l":                true,
	"m":                true,
	"n":                true,
	"o":                true,
	"p":                true,
	"q":                true,
	"r":                true,
	"s":                true,
	"t":                true,
	"u":                true,
	"v":                true,
	"w":                true,
	"x":                true,
	"y":                true,
	"z":                true,
	"braceleft":        true,
	"bar":              true,
	"braceright":       true,
	"asciitilde":       true,
	"Adieresis":        true,
	"Aring":            true,
	"Ccedilla":         true,
	"Eacute":           true,
	"Ntilde":           true,
	"Odieresis":        true,
	"Udieresis":        true,
	"aacute":           true,
	"agrave":           true,
	"acircumflex":      true,
	"adieresis":        true,
	"atilde":           true,
	"aring":            true,
	"ccedilla":         true,
	"eacute":           true,
	"egrave":           true,
	"ecircumflex":      true,
	"edieresis":        true,
	"iacute":           true,
	"igrave":           true,
	"icircumflex":      true,
	"idieresis":        true,
	"ntilde":           true,
	"oacute":           true,
	"ograve":           true,
	"ocircumflex":      true,
	"odieresis":        true,
	"otilde":           true,
	"uacute":           true,
	"ugrave":           true,
	"ucircumflex":      true,
	"udieresis":        true,
	"dagger":           true,
	"degree":           true,
	"cent":             true,
	"sterling":         true,
	"section":          true,
	"bullet":           true,
	"paragraph":        true,
	"germandbls":       true,
	"registered":       true,
	"copyright":        true,
	"trademark":        true,
	"acute":            true,
	"dieresis":         true,
	"notequal":         true,
	"AE":               true,
	"Oslash":           true,
	"infinity":         true,
	"plusminus":        true,
	"lessequal":        true,
	"greaterequal":     true,
	"yen":              true,
	"mu":               true,
	"partialdiff":      true,
	"summation":        true,
	"product":          true,
	"pi":               true,
	"integral":         true,
	"ordfeminine":      true,
	"ordmasculine":     true,
	"Omega":            true,
	"ae":               true,
	"oslash":           true,
	"questiondown":     true,
	"exclamdown":       true,
	"logicalnot":       true,
	"radical":          true,
	"florin":           true,
	"approxequal":      true,
	"Delta":            true,
	"guillemotleft":    true,
	"guillemotright":   true,
	"ellipsis":         true,
	"nonbreakingspace": true,
	"Agrave":           true,
	"Atilde":           true,
	"Otilde":           true,
	"OE":               true,
	"oe":               true,
	"endash":           true,
	"emdash":           true,
	"quotedblleft":     true,
	"quotedblright":    true,
	"quoteleft":        true,
	"quoteright":       true,
	"divide":           true,
	"lozenge":          true,
	"ydieresis":        true,
	"Ydieresis":        true,
	"fraction":         true,
	"currency":         true,
	"guilsinglleft":    true,
	"guilsinglright":   true,
	"fi":               true,
	"fl":               true,
	"daggerdbl":        true,
	"periodcentered":   true,
	"quotesinglbase":   true,
	"quotedblbase":     true,
	"perthousand":      true,
	"Acircumflex":      true,
	"Ecircumflex":      true,
	"Aacute":           true,
	"Edieresis":        true,
	"Egrave":           true,
	"Iacute":           true,
	"Icircumflex":      true,
	"Idieresis":        true,
	"Igrave":           true,
	"Oacute":           true,
	"Ocircumflex":      true,
	"apple":            true,
	"Ograve":           true,
	"Uacute":           true,
	"Ucircumflex":      true,
	"Ugrave":           true,
	"dotlessi":         true,
	"circumflex":       true,
	"tilde":            true,
	"macron":           true,
	"breve":            true,
	"dotaccent":        true,
	"ring":             true,
	"cedilla":          true,
	"hungarumlaut":     true,
	"ogonek":           true,
	"caron":            true,
	"Lslash":           true,
	"lslash":           true,
	"Scaron":           true,
	"scaron":           true,
	"Zcaron":           true,
	"zcaron":           true,
	"brokenbar":        true,
	"Eth":              true,
	"eth":              true,
	"Yacute":           true,
	"yacute":           true,
	"Thorn":            true,
	"thorn":            true,
	"minus":            true,
	"multiply":         true,
	"onesuperior":      true,
	"twosuperior":      true,
	"threesuperior":    true,
	"onehalf":          true,
	"onequarter":       true,
	"threequarters":    true,
	"franc":            true,
	"Gbreve":           true,
	"gbreve":           true,
	"Idotaccent":       true,
	"Scedilla":         true,
	"scedilla":         true,
	"Cacute":           true,
	"cacute":           true,
	"Ccaron":           true,
	"ccaron":           true,
	"dcroat":           true,
}

// aglfn comes from
// https://raw.githubusercontent.com/adobe-type-tools/agl-aglfn/master/aglfn.txt
var aglfn = map[rune]string{
	0x0020: "space",
	0x0021: "exclam",
	0x0022: "quotedbl",
	0x0023: "numbersign",
	0x0024: "dollar",
	0x0025: "percent",
	0x0026: "ampersand",
	0x0027: "quotesingle",
	0x0028: "parenleft",
	0x0029: "parenright",
	0x002a: "asterisk",
	0x002b: "plus",
	0x002c: "comma",
	0x002d: "hyphen",
	0x002e: "period",
	0x002f: "slash",
	0x0030: "zero",
	0x0031: "one",
	0x0032: "two",
	0x0033: "three",
	0x0034: "four",
	0x0035: "five",
	0x0036: "six",
	0x0037: "seven",
	0x0038: "eight",
	0x0039: "nine",
	0x003a: "colon",
	0x003b: "semicolon",
	0x003c: "less",
	0x003d: "equal",
	0x003e: "greater",
	0x003f: "question",
	0x0040: "at",
	0x0041: "A",
	0x0042: "B",
	0x0043: "C",
	0x0044: "D",
	0x0045: "E",
	0x0046: "F",
	0x0047: "G",
	0x0048: "H",
	0x0049: "I",
	0x004a: "J",
	0x004b: "K",
	0x004c: "L",
	0x004d: "M",
	0x004e: "N",
	0x004f: "O",
	0x0050: "P",
	0x0051: "Q",
	0x0052: "R",
	0x0053: "S",
	0x0054: "T",
	0x0055: "U",
	0x0056: "V",
	0x0057: "W",
	0x0058: "X",
	0x0059: "Y",
	0x005a: "Z",
	0x005b: "bracketleft",
	0x005c: "backslash",
	0x005d: "bracketright",
	0x005e: "asciicircum",
	0x005f: "underscore",
	0x0060: "grave",
	0x0061: "a",
	0x0062: "b",
	0x0063: "c",
	0x0064: "d",
	0x0065: "e",
	0x0066: "f",
	0x0067: "g",
	0x0068: "h",
	0x0069: "i",
	0x006a: "j",
	0x006b: "k",
	0x006c: "l",
	0x006d: "m",
	0x006e: "n",
	0x006f: "o",
	0x0070: "p",
	0x0071: "q",
	0x0072: "r",
	0x0073: "s",
	0x0074: "t",
	0x0075: "u",
	0x0076: "v",
	0x0077: "w",
	0x0078: "x",
	0x0079: "y",
	0x007a: "z",
	0x007b: "braceleft",
	0x007c: "bar",
	0x007d: "braceright",
	0x007e: "asciitilde",
	0x00a1: "exclamdown",
	0x00a2: "cent",
	0x00a3: "sterling",
	0x00a4: "currency",
	0x00a5: "yen",
	0x00a6: "brokenbar",
	0x00a7: "section",
	0x00a8: "dieresis",
	0x00a9: "copyright",
	0x00aa: "ordfeminine",
	0x00ab: "guillemotleft",
	0x00ac: "logicalnot",
	0x00ae: "registered",
	0x00af: "macron",
	0x00b0: "degree",
	0x00b1: "plusminus",
	0x00b4: "acute",
	0x00b5: "mu",
	0x00b6: "paragraph",
	0x00b7: "periodcentered",
	0x00b8: "cedilla",
	0x00ba: "ordmasculine",
	0x00bb: "guillemotright",
	0x00bc: "onequarter",
	0x00bd: "onehalf",
	0x00be: "threequarters",
	0x00bf: "questiondown",
	0x00c0: "Agrave",
	0x00c1: "Aacute",
	0x00c2: "Acircumflex",
	0x00c3: "Atilde",
	0x00c4: "Adieresis",
	0x00c5: "Aring",
	0x00c6: "AE",
	0x00c7: "Ccedilla",
	0x00c8: "Egrave",
	0x00c9: "Eacute",
	0x00ca: "Ecircumflex",
	0x00cb: "Edieresis",
	0x00cc: "Igrave",
	0x00cd: "Iacute",
	0x00ce: "Icircumflex",
	0x00cf: "Idieresis",
	0x00d0: "Eth",
	0x00d1: "Ntilde",
	0x00d2: "Ograve",
	0x00d3: "Oacute",
	0x00d4: "Ocircumflex",
	0x00d5: "Otilde",
	0x00d6: "Odieresis",
	0x00d7: "multiply",
	0x00d8: "Oslash",
	0x00d9: "Ugrave",
	0x00da: "Uacute",
	0x00db: "Ucircumflex",
	0x00dc: "Udieresis",
	0x00dd: "Yacute",
	0x00de: "Thorn",
	0x00df: "germandbls",
	0x00e0: "agrave",
	0x00e1: "aacute",
	0x00e2: "acircumflex",
	0x00e3: "atilde",
	0x00e4: "adieresis",
	0x00e5: "aring",
	0x00e6: "ae",
	0x00e7: "ccedilla",
	0x00e8: "egrave",
	0x00e9: "eacute",
	0x00ea: "ecircumflex",
	0x00eb: "edieresis",
	0x00ec: "igrave",
	0x00ed: "iacute",
	0x00ee: "icircumflex",
	0x00ef: "idieresis",
	0x00f0: "eth",
	0x00f1: "ntilde",
	0x00f2: "ograve",
	0x00f3: "oacute",
	0x00f4: "ocircumflex",
	0x00f5: "otilde",
	0x00f6: "odieresis",
	0x00f7: "divide",
	0x00f8: "oslash",
	0x00f9: "ugrave",
	0x00fa: "uacute",
	0x00fb: "ucircumflex",
	0x00fc: "udieresis",
	0x00fd: "yacute",
	0x00fe: "thorn",
	0x00ff: "ydieresis",
	0x0100: "Amacron",
	0x0101: "amacron",
	0x0102: "Abreve",
	0x0103: "abreve",
	0x0104: "Aogonek",
	0x0105: "aogonek",
	0x0106: "Cacute",
	0x0107: "cacute",
	0x0108: "Ccircumflex",
	0x0109: "ccircumflex",
	0x010a: "Cdotaccent",
	0x010b: "cdotaccent",
	0x010c: "Ccaron",
	0x010d: "ccaron",
	0x010e: "Dcaron",
	0x010f: "dcaron",
	0x0110: "Dcroat",
	0x0111: "dcroat",
	0x0112: "Emacron",
	0x0113: "emacron",
	0x0114: "Ebreve",
	0x0115: "ebreve",
	0x0116: "Edotaccent",
	0x0117: "edotaccent",
	0x0118: "Eogonek",
	0x0119: "eogonek",
	0x011a: "Ecaron",
	0x011b: "ecaron",
	0x011c: "Gcircumflex",
	0x011d: "gcircumflex",
	0x011e: "Gbreve",
	0x011f: "gbreve",
	0x0120: "Gdotaccent",
	0x0121: "gdotaccent",
	0x0124: "Hcircumflex",
	0x0125: "hcircumflex",
	0x0126: "Hbar",
	0x0127: "hbar",
	0x0128: "Itilde",
	0x0129: "itilde",
	0x012a: "Imacron",
	0x012b: "imacron",
	0x012c: "Ibreve",
	0x012d: "ibreve",
	0x012e: "Iogonek",
	0x012f: "iogonek",
	0x0130: "Idotaccent",
	0x0131: "dotlessi",
	0x0132: "IJ",
	0x0133: "ij",
	0x0134: "Jcircumflex",
	0x0135: "jcircumflex",
	0x0138: "kgreenlandic",
	0x0139: "Lacute",
	0x013a: "lacute",
	0x013d: "Lcaron",
	0x013e: "lcaron",
	0x013f: "Ldot",
	0x0140: "ldot",
	0x0141: "Lslash",
	0x0142: "lslash",
	0x0143: "Nacute",
	0x0144: "nacute",
	0x0147: "Ncaron",
	0x0148: "ncaron",
	0x0149: "napostrophe",
	0x014a: "Eng",
	0x014b: "eng",
	0x014c: "Omacron",
	0x014d: "omacron",
	0x014e: "Obreve",
	0x014f: "obreve",
	0x0150: "Ohungarumlaut",
	0x0151: "ohungarumlaut",
	0x0152: "OE",
	0x0153: "oe",
	0x0154: "Racute",
	0x0155: "racute",
	0x0158: "Rcaron",
	0x0159: "rcaron",
	0x015a: "Sacute",
	0x015b: "sacute",
	0x015c: "Scircumflex",
	0x015d: "scircumflex",
	0x015e: "Scedilla",
	0x015f: "scedilla",
	0x0160: "Scaron",
	0x0161: "scaron",
	0x0164: "Tcaron",
	0x0165: "tcaron",
	0x0166: "Tbar",
	0x0167: "tbar",
	0x0168: "Utilde",
	0x0169: "utilde",
	0x016a: "Umacron",
	0x016b: "umacron",
	0x016c: "Ubreve",
	0x016d: "ubreve",
	0x016e: "Uring",
	0x016f: "uring",
	0x0170: "Uhungarumlaut",
	0x0171: "uhungarumlaut",
	0x0172: "Uogonek",
	0x0173: "uogonek",
	0x0174: "Wcircumflex",
	0x0175: "wcircumflex",
	0x0176: "Ycircumflex",
	0x0177: "ycircumflex",
	0x0178: "Ydieresis",
	0x0179: "Zacute",
	0x017a: "zacute",
	0x017b: "Zdotaccent",
	0x017c: "zdotaccent",
	0x017d: "Zcaron",
	0x017e: "zcaron",
	0x017f: "longs",
	0x0192: "florin",
	0x01a0: "Ohorn",
	0x01a1: "ohorn",
	0x01af: "Uhorn",
	0x01b0: "uhorn",
	0x01e6: "Gcaron",
	0x01e7: "gcaron",
	0x01fa: "Aringacute",
	0x01fb: "aringacute",
	0x01fc: "AEacute",
	0x01fd: "aeacute",
	0x01fe: "Oslashacute",
	0x01ff: "oslashacute",
	0x02c6: "circumflex",
	0x02c7: "caron",
	0x02d8: "breve",
	0x02d9: "dotaccent",
	0x02da: "ring",
	0x02db: "ogonek",
	0x02dc: "tilde",
	0x02dd: "hungarumlaut",
	0x0300: "gravecomb",
	0x0301: "acutecomb",
	0x0303: "tildecomb",
	0x0309: "hookabovecomb",
	0x0323: "dotbelowcomb",
	0x0384: "tonos",
	0x0385: "dieresistonos",
	0x0386: "Alphatonos",
	0x0387: "anoteleia",
	0x0388: "Epsilontonos",
	0x0389: "Etatonos",
	0x038a: "Iotatonos",
	0x038c: "Omicrontonos",
	0x038e: "Upsilontonos",
	0x038f: "Omegatonos",
	0x0390: "iotadieresistonos",
	0x0391: "Alpha",
	0x0392: "Beta",
	0x0393: "Gamma",
	0x0395: "Epsilon",
	0x0396: "Zeta",
	0x0397: "Eta",
	0x0398: "Theta",
	0x0399: "Iota",
	0x039a: "Kappa",
	0x039b: "Lambda",
	0x039c: "Mu",
	0x039d: "Nu",
	0x039e: "Xi",
	0x039f: "Omicron",
	0x03a0: "Pi",
	0x03a1: "Rho",
	0x03a3: "Sigma",
	0x03a4: "Tau",
	0x03a5: "Upsilon",
	0x03a6: "Phi",
	0x03a7: "Chi",
	0x03a8: "Psi",
	0x03aa: "Iotadieresis",
	0x03ab: "Upsilondieresis",
	0x03ac: "alphatonos",
	0x03ad: "epsilontonos",
	0x03ae: "etatonos",
	0x03af: "iotatonos",
	0x03b0: "upsilondieresistonos",
	0x03b1: "alpha",
	0x03b2: "beta",
	0x03b3: "gamma",
	0x03b4: "delta",
	0x03b5: "epsilon",
	0x03b6: "zeta",
	0x03b7: "eta",
	0x03b8: "theta",
	0x03b9: "iota",
	0x03ba: "kappa",
	0x03bb: "lambda",
	0x03bd: "nu",
	0x03be: "xi",
	0x03bf: "omicron",
	0x03c0: "pi",
	0x03c1: "rho",
	0x03c2: "sigma1",
	0x03c3: "sigma",
	0x03c4: "tau",
	0x03c5: "upsilon",
	0x03c6: "phi",
	0x03c7: "chi",
	0x03c8: "psi",
	0x03c9: "omega",
	0x03ca: "iotadieresis",
	0x03cb: "upsilondieresis",
	0x03cc: "omicrontonos",
	0x03cd: "upsilontonos",
	0x03ce: "omegatonos",
	0x03d1: "theta1",
	0x03d2: "Upsilon1",
	0x03d5: "phi1",
	0x03d6: "omega1",
	0x1e80: "Wgrave",
	0x1e81: "wgrave",
	0x1e82: "Wacute",
	0x1e83: "wacute",
	0x1e84: "Wdieresis",
	0x1e85: "wdieresis",
	0x1ef2: "Ygrave",
	0x1ef3: "ygrave",
	0x2012: "figuredash",
	0x2013: "endash",
	0x2014: "emdash",
	0x2017: "underscoredbl",
	0x2018: "quoteleft",
	0x2019: "quoteright",
	0x201a: "quotesinglbase",
	0x201b: "quotereversed",
	0x201c: "quotedblleft",
	0x201d: "quotedblright",
	0x201e: "quotedblbase",
	0x2020: "dagger",
	0x2021: "daggerdbl",
	0x2022: "bullet",
	0x2024: "onedotenleader",
	0x2025: "twodotenleader",
	0x2026: "ellipsis",
	0x2030: "perthousand",
	0x2032: "minute",
	0x2033: "second",
	0x2039: "guilsinglleft",
	0x203a: "guilsinglright",
	0x203c: "exclamdbl",
	0x2044: "fraction",
	0x20a1: "colonmonetary",
	0x20a3: "franc",
	0x20a4: "lira",
	0x20a7: "peseta",
	0x20ab: "dong",
	0x20ac: "Euro",
	0x2111: "Ifraktur",
	0x2118: "weierstrass",
	0x211c: "Rfraktur",
	0x211e: "prescription",
	0x2122: "trademark",
	0x2126: "Omega",
	0x212e: "estimated",
	0x2135: "aleph",
	0x2153: "onethird",
	0x2154: "twothirds",
	0x215b: "oneeighth",
	0x215c: "threeeighths",
	0x215d: "fiveeighths",
	0x215e: "seveneighths",
	0x2190: "arrowleft",
	0x2191: "arrowup",
	0x2192: "arrowright",
	0x2193: "arrowdown",
	0x2194: "arrowboth",
	0x2195: "arrowupdn",
	0x21a8: "arrowupdnbse",
	0x21b5: "carriagereturn",
	0x21d0: "arrowdblleft",
	0x21d1: "arrowdblup",
	0x21d2: "arrowdblright",
	0x21d3: "arrowdbldown",
	0x21d4: "arrowdblboth",
	0x2200: "universal",
	0x2202: "partialdiff",
	0x2203: "existential",
	0x2205: "emptyset",
	0x2206: "Delta",
	0x2207: "gradient",
	0x2208: "element",
	0x2209: "notelement",
	0x220b: "suchthat",
	0x220f: "product",
	0x2211: "summation",
	0x2212: "minus",
	0x2217: "asteriskmath",
	0x221a: "radical",
	0x221d: "proportional",
	0x221e: "infinity",
	0x221f: "orthogonal",
	0x2220: "angle",
	0x2227: "logicaland",
	0x2228: "logicalor",
	0x2229: "intersection",
	0x222a: "union",
	0x222b: "integral",
	0x2234: "therefore",
	0x223c: "similar",
	0x2245: "congruent",
	0x2248: "approxequal",
	0x2260: "notequal",
	0x2261: "equivalence",
	0x2264: "lessequal",
	0x2265: "greaterequal",
	0x2282: "propersubset",
	0x2283: "propersuperset",
	0x2284: "notsubset",
	0x2286: "reflexsubset",
	0x2287: "reflexsuperset",
	0x2295: "circleplus",
	0x2297: "circlemultiply",
	0x22a5: "perpendicular",
	0x22c5: "dotmath",
	0x2302: "house",
	0x2310: "revlogicalnot",
	0x2320: "integraltp",
	0x2321: "integralbt",
	0x2329: "angleleft",
	0x232a: "angleright",
	0x2500: "SF100000",
	0x2502: "SF110000",
	0x250c: "SF010000",
	0x2510: "SF030000",
	0x2514: "SF020000",
	0x2518: "SF040000",
	0x251c: "SF080000",
	0x2524: "SF090000",
	0x252c: "SF060000",
	0x2534: "SF070000",
	0x253c: "SF050000",
	0x2550: "SF430000",
	0x2551: "SF240000",
	0x2552: "SF510000",
	0x2553: "SF520000",
	0x2554: "SF390000",
	0x2555: "SF220000",
	0x2556: "SF210000",
	0x2557: "SF250000",
	0x2558: "SF500000",
	0x2559: "SF490000",
	0x255a: "SF380000",
	0x255b: "SF280000",
	0x255c: "SF270000",
	0x255d: "SF260000",
	0x255e: "SF360000",
	0x255f: "SF370000",
	0x2560: "SF420000",
	0x2561: "SF190000",
	0x2562: "SF200000",
	0x2563: "SF230000",
	0x2564: "SF470000",
	0x2565: "SF480000",
	0x2566: "SF410000",
	0x2567: "SF450000",
	0x2568: "SF460000",
	0x2569: "SF400000",
	0x256a: "SF540000",
	0x256b: "SF530000",
	0x256c: "SF440000",
	0x2580: "upblock",
	0x2584: "dnblock",
	0x2588: "block",
	0x258c: "lfblock",
	0x2590: "rtblock",
	0x2591: "ltshade",
	0x2592: "shade",
	0x2593: "dkshade",
	0x25a0: "filledbox",
	0x25a1: "H22073",
	0x25aa: "H18543",
	0x25ab: "H18551",
	0x25ac: "filledrect",
	0x25b2: "triagup",
	0x25ba: "triagrt",
	0x25bc: "triagdn",
	0x25c4: "triaglf",
	0x25ca: "lozenge",
	0x25cb: "circle",
	0x25cf: "H18533",
	0x25d8: "invbullet",
	0x25d9: "invcircle",
	0x25e6: "openbullet",
	0x263a: "smileface",
	0x263b: "invsmileface",
	0x263c: "sun",
	0x2640: "female",
	0x2642: "male",
	0x2660: "spade",
	0x2663: "club",
	0x2665: "heart",
	0x2666: "diamond",
	0x266a: "musicalnote",
	0x266b: "musicalnotedbl",
}
