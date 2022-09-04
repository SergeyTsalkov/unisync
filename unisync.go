package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')

	fmt.Printf("%#v", line)
	handle(line)
}

func handle(line string) {
	words := strings.Fields(line)
	cmd := strings.ToUpper(words[0])

	switch cmd {
	case "REQFILELIST":
		return handleREQFILELIST()
	}

	fmt.Printf("%#v", cmd)
}

func handleREQFILELIST() {
	files, _ := os.ReadDir(".")

}
