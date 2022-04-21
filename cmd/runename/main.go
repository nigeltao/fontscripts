// runename prints the official names for Unicode code points.
//
// $ runename 5c 2603 1f574
// U+0000005C REVERSE SOLIDUS
// U+00002603 SNOWMAN
// U+0001F574 MAN IN BUSINESS SUIT LEVITATING
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/runenames"
)

func main() {
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "U+") || strings.HasPrefix(a, "u+") {
			a = a[2:]
		}
		name := ""
		n, err := strconv.ParseInt(a, 16, 64)
		if err != nil {
			name = "!" + err.Error()
		} else {
			name = runenames.Name(rune(n))
		}

		fmt.Printf("U+%08X %s\n", n, name)
	}
}
