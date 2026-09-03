package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	var input string
	if len(os.Args) > 1 {
		input = strings.Join(os.Args[1:], " ")
	} else {
		data, err := readStdin()
		if err != nil {
			fmt.Fprintln(os.Stderr, "connfmt: reading stdin:", err)
			os.Exit(1)
		}
		input = data
	}

	out, err := Normalize(input)
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
