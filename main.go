package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	redact := flag.Bool("redact", false, "mask password values in the output")
	flag.Parse()

	var input string
	if flag.NArg() > 0 {
		input = strings.Join(flag.Args(), " ")
	} else {
		data, err := readStdin()
		if err != nil {
			fmt.Fprintln(os.Stderr, "connfmt: reading stdin:", err)
			os.Exit(1)
		}
		input = data
	}

	out, err := Normalize(input, *redact)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connfmt:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

func readStdin() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var b strings.Builder
	for scanner.Scan() {
		b.WriteString(scanner.Text())
		b.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(b.String()), nil
}
