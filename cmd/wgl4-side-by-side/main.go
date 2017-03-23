// wgl4-side-by-side prints the glyphs of the WGL-4 repertoire from the given
// TrueType fonts.
package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/runenames"
)

const (
	width       = 64
	largeHeight = 64
	smallHeight = 32
	hinting     = font.HintingNone
)

var (
	names      []string
	largeFaces []font.Face
	smallFaces []font.Face

	goregularFont      *truetype.Font
	goregularSmallFace font.Face
	goregularTinyFace  font.Face
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s filename1.ttf filename2.ttf filename3.ttf etc\n", os.Args[0])
		os.Exit(1)
	}

	var err error
	goregularFont, err = truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	goregularSmallFace = truetype.NewFace(goregularFont, &truetype.Options{
		Size:    24,
		Hinting: hinting,
	})
	goregularTinyFace = truetype.NewFace(goregularFont, &truetype.Options{
		Size:    10,
		Hinting: hinting,
	})

	names = make([]string, len(args))
	largeFaces = make([]font.Face, len(args))
	smallFaces = make([]font.Face, len(args))
	for i, arg := range args {
		fontBytes, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Fatal(err)
		}
		f, err := truetype.Parse(fontBytes)
		if err != nil {
			log.Fatal(err)
		}
		names[i] = fmt.Sprintf("%s; %s",
			f.Name(truetype.NameIDFontFullName),
			f.Name(truetype.NameIDNameTableVersion),
		)
		largeFaces[i] = truetype.NewFace(f, &truetype.Options{
			Size:    48,
			Hinting: hinting,
		})
		smallFaces[i] = truetype.NewFace(f, &truetype.Options{
			Size:    24,
			Hinting: hinting,
		})
	}

	cuts := []rune{
		0x0080,
		0x0100,
		0x0200,
		0x0400,
		0x1000,
		0x2500,
		0xFFFF,
	}
	prevCut := rune(0)
	for _, cut := range cuts {
		do(prevCut, cut)
		prevCut = cut
	}
}

func do(lo, hi rune) {
	const yMax = 9216

	dst := image.NewRGBA(image.Rect(0, 0, width*len(names)+384, yMax+smallHeight*(1+len(names))))
	bounds := dst.Bounds()
	draw.Draw(dst, bounds, image.White, image.Point{}, draw.Src)

	for y := 0; y < yMax; y++ {
		for j := range largeFaces {
			dst.SetRGBA(width*j+16, y, color.RGBA{0xe0, 0xe0, 0xe0, 0xff})
		}
	}

	gray := image.NewUniform(color.RGBA{0x80, 0x80, 0x80, 0xff})
	d := &font.Drawer{
		Dst: dst,
	}
	y := 64
	for _, c := range repertoire {
		if c < lo || hi <= c {
			continue
		}

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.SetRGBA(x, y, color.RGBA{0xe0, 0xe0, 0xe0, 0xff})
		}

		d.Src = gray
		d.Face = goregularTinyFace
		d.Dot = fixed.P(width*len(names)+64, y+12)
		d.DrawString(runenames.Name(rune(c)))

		d.Src = image.Black
		d.Face = goregularSmallFace
		d.Dot = fixed.P(width*len(names)+64, y)
		d.DrawString(fmt.Sprintf("U+%04X", c))

		s := string(c)
		for j, face := range largeFaces {
			d.Face = face
			d.Dot = fixed.P(width*j+16, y)
			d.DrawString(s)
		}
		y += largeHeight
	}

	for i, s := range names {
		d.Face = smallFaces[i]
		d.Dot = fixed.P(16, yMax+(i+1)*smallHeight)
		d.DrawString(s)
	}

	filename := fmt.Sprintf("side-by-side-%04x-%04x.png", lo, hi)
	outFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, dst)
	if err != nil {
		log.Fatal(err)
	}
	err = b.Flush()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote %s\n", filename)
}

// repertoire is the WGL-4 repertoire plus:
//	- U+F800 <Private Use>
//	- U+FFFD REPLACEMENT CHARACTER
var repertoire = []rune{
	'\u0021', '\u0022', '\u0023', '\u0024', '\u0025', '\u0026', '\u0027', '\u0028',
	'\u0029', '\u002a', '\u002b', '\u002c', '\u002d', '\u002e', '\u002f', '\u0030',
	'\u0031', '\u0032', '\u0033', '\u0034', '\u0035', '\u0036', '\u0037', '\u0038',
	'\u0039', '\u003a', '\u003b', '\u003c', '\u003d', '\u003e', '\u003f', '\u0040',
	'\u0041', '\u0042', '\u0043', '\u0044', '\u0045', '\u0046', '\u0047', '\u0048',
	'\u0049', '\u004a', '\u004b', '\u004c', '\u004d', '\u004e', '\u004f', '\u0050',
	'\u0051', '\u0052', '\u0053', '\u0054', '\u0055', '\u0056', '\u0057', '\u0058',
	'\u0059', '\u005a', '\u005b', '\u005c', '\u005d', '\u005e', '\u005f', '\u0060',
	'\u0061', '\u0062', '\u0063', '\u0064', '\u0065', '\u0066', '\u0067', '\u0068',
	'\u0069', '\u006a', '\u006b', '\u006c', '\u006d', '\u006e', '\u006f', '\u0070',
	'\u0071', '\u0072', '\u0073', '\u0074', '\u0075', '\u0076', '\u0077', '\u0078',
	'\u0079', '\u007a', '\u007b', '\u007c', '\u007d', '\u007e', '\u00a1', '\u00a2',
	'\u00a3', '\u00a4', '\u00a5', '\u00a6', '\u00a7', '\u00a8', '\u00a9', '\u00aa',
	'\u00ab', '\u00ac', '\u00ae', '\u00af', '\u00b0', '\u00b1', '\u00b2', '\u00b3',
	'\u00b4', '\u00b5', '\u00b6', '\u00b7', '\u00b8', '\u00b9', '\u00ba', '\u00bb',
	'\u00bc', '\u00bd', '\u00be', '\u00bf', '\u00c0', '\u00c1', '\u00c2', '\u00c3',
	'\u00c4', '\u00c5', '\u00c6', '\u00c7', '\u00c8', '\u00c9', '\u00ca', '\u00cb',
	'\u00cc', '\u00cd', '\u00ce', '\u00cf', '\u00d0', '\u00d1', '\u00d2', '\u00d3',
	'\u00d4', '\u00d5', '\u00d6', '\u00d7', '\u00d8', '\u00d9', '\u00da', '\u00db',
	'\u00dc', '\u00dd', '\u00de', '\u00df', '\u00e0', '\u00e1', '\u00e2', '\u00e3',
	'\u00e4', '\u00e5', '\u00e6', '\u00e7', '\u00e8', '\u00e9', '\u00ea', '\u00eb',
	'\u00ec', '\u00ed', '\u00ee', '\u00ef', '\u00f0', '\u00f1', '\u00f2', '\u00f3',
	'\u00f4', '\u00f5', '\u00f6', '\u00f7', '\u00f8', '\u00f9', '\u00fa', '\u00fb',
	'\u00fc', '\u00fd', '\u00fe', '\u00ff', '\u0100', '\u0101', '\u0102', '\u0103',
	'\u0104', '\u0105', '\u0106', '\u0107', '\u0108', '\u0109', '\u010a', '\u010b',
	'\u010c', '\u010d', '\u010e', '\u010f', '\u0110', '\u0111', '\u0112', '\u0113',
	'\u0114', '\u0115', '\u0116', '\u0117', '\u0118', '\u0119', '\u011a', '\u011b',
	'\u011c', '\u011d', '\u011e', '\u011f', '\u0120', '\u0121', '\u0122', '\u0123',
	'\u0124', '\u0125', '\u0126', '\u0127', '\u0128', '\u0129', '\u012a', '\u012b',
	'\u012c', '\u012d', '\u012e', '\u012f', '\u0130', '\u0131', '\u0132', '\u0133',
	'\u0134', '\u0135', '\u0136', '\u0137', '\u0138', '\u0139', '\u013a', '\u013b',
	'\u013c', '\u013d', '\u013e', '\u013f', '\u0140', '\u0141', '\u0142', '\u0143',
	'\u0144', '\u0145', '\u0146', '\u0147', '\u0148', '\u0149', '\u014a', '\u014b',
	'\u014c', '\u014d', '\u014e', '\u014f', '\u0150', '\u0151', '\u0152', '\u0153',
	'\u0154', '\u0155', '\u0156', '\u0157', '\u0158', '\u0159', '\u015a', '\u015b',
	'\u015c', '\u015d', '\u015e', '\u015f', '\u0160', '\u0161', '\u0162', '\u0163',
	'\u0164', '\u0165', '\u0166', '\u0167', '\u0168', '\u0169', '\u016a', '\u016b',
	'\u016c', '\u016d', '\u016e', '\u016f', '\u0170', '\u0171', '\u0172', '\u0173',
	'\u0174', '\u0175', '\u0176', '\u0177', '\u0178', '\u0179', '\u017a', '\u017b',
	'\u017c', '\u017d', '\u017e', '\u017f', '\u0192', '\u01fa', '\u01fb', '\u01fc',
	'\u01fd', '\u01fe', '\u01ff', '\u02c6', '\u02c7', '\u02c9', '\u02d8', '\u02d9',
	'\u02da', '\u02db', '\u02dc', '\u02dd', '\u0384', '\u0385', '\u0386', '\u0387',
	'\u0388', '\u0389', '\u038a', '\u038c', '\u038e', '\u038f', '\u0390', '\u0391',
	'\u0392', '\u0393', '\u0394', '\u0395', '\u0396', '\u0397', '\u0398', '\u0399',
	'\u039a', '\u039b', '\u039c', '\u039d', '\u039e', '\u039f', '\u03a0', '\u03a1',
	'\u03a3', '\u03a4', '\u03a5', '\u03a6', '\u03a7', '\u03a8', '\u03a9', '\u03aa',
	'\u03ab', '\u03ac', '\u03ad', '\u03ae', '\u03af', '\u03b0', '\u03b1', '\u03b2',
	'\u03b3', '\u03b4', '\u03b5', '\u03b6', '\u03b7', '\u03b8', '\u03b9', '\u03ba',
	'\u03bb', '\u03bc', '\u03bd', '\u03be', '\u03bf', '\u03c0', '\u03c1', '\u03c2',
	'\u03c3', '\u03c4', '\u03c5', '\u03c6', '\u03c7', '\u03c8', '\u03c9', '\u03ca',
	'\u03cb', '\u03cc', '\u03cd', '\u03ce', '\u0400', '\u0401', '\u0402', '\u0403',
	'\u0404', '\u0405', '\u0406', '\u0407', '\u0408', '\u0409', '\u040a', '\u040b',
	'\u040c', '\u040d', '\u040e', '\u040f', '\u0410', '\u0411', '\u0412', '\u0413',
	'\u0414', '\u0415', '\u0416', '\u0417', '\u0418', '\u0419', '\u041a', '\u041b',
	'\u041c', '\u041d', '\u041e', '\u041f', '\u0420', '\u0421', '\u0422', '\u0423',
	'\u0424', '\u0425', '\u0426', '\u0427', '\u0428', '\u0429', '\u042a', '\u042b',
	'\u042c', '\u042d', '\u042e', '\u042f', '\u0430', '\u0431', '\u0432', '\u0433',
	'\u0434', '\u0435', '\u0436', '\u0437', '\u0438', '\u0439', '\u043a', '\u043b',
	'\u043c', '\u043d', '\u043e', '\u043f', '\u0440', '\u0441', '\u0442', '\u0443',
	'\u0444', '\u0445', '\u0446', '\u0447', '\u0448', '\u0449', '\u044a', '\u044b',
	'\u044c', '\u044d', '\u044e', '\u044f', '\u0450', '\u0451', '\u0452', '\u0453',
	'\u0454', '\u0455', '\u0456', '\u0457', '\u0458', '\u0459', '\u045a', '\u045b',
	'\u045c', '\u045d', '\u045e', '\u045f', '\u0490', '\u0491', '\u1e80', '\u1e81',
	'\u1e82', '\u1e83', '\u1e84', '\u1e85', '\u1ef2', '\u1ef3', '\u2013', '\u2014',
	'\u2015', '\u2017', '\u2018', '\u2019', '\u201a', '\u201b', '\u201c', '\u201d',
	'\u201e', '\u2020', '\u2021', '\u2022', '\u2026', '\u2030', '\u2032', '\u2033',
	'\u2039', '\u203a', '\u203c', '\u203e', '\u2044', '\u207f', '\u20a3', '\u20a4',
	'\u20a7', '\u20ac', '\u2105', '\u2113', '\u2116', '\u2122', '\u212e', '\u215b',
	'\u215c', '\u215d', '\u215e', '\u2190', '\u2191', '\u2192', '\u2193', '\u2194',
	'\u2195', '\u21a8', '\u2202', '\u2206', '\u220f', '\u2211', '\u2212', '\u2215',
	'\u2219', '\u221a', '\u221e', '\u221f', '\u2229', '\u222b', '\u2248', '\u2260',
	'\u2261', '\u2264', '\u2265', '\u2302', '\u2310', '\u2320', '\u2321', '\u2500',
	'\u2502', '\u250c', '\u2510', '\u2514', '\u2518', '\u251c', '\u2524', '\u252c',
	'\u2534', '\u253c', '\u2550', '\u2551', '\u2552', '\u2553', '\u2554', '\u2555',
	'\u2556', '\u2557', '\u2558', '\u2559', '\u255a', '\u255b', '\u255c', '\u255d',
	'\u255e', '\u255f', '\u2560', '\u2561', '\u2562', '\u2563', '\u2564', '\u2565',
	'\u2566', '\u2567', '\u2568', '\u2569', '\u256a', '\u256b', '\u256c', '\u2580',
	'\u2584', '\u2588', '\u258c', '\u2590', '\u2591', '\u2592', '\u2593', '\u25a0',
	'\u25a1', '\u25aa', '\u25ab', '\u25ac', '\u25b2', '\u25ba', '\u25bc', '\u25c4',
	'\u25ca', '\u25cb', '\u25cf', '\u25d8', '\u25d9', '\u25e6', '\u263a', '\u263b',
	'\u263c', '\u2640', '\u2642', '\u2660', '\u2663', '\u2665', '\u2666', '\u266a',
	'\u266b', '\uf800', '\ufb01', '\ufb02', '\ufffd',
}
