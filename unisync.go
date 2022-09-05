package main

import (
	"fmt"

	"unisync/filelist"
)

func main() {
	// reader := bufio.NewReader(os.Stdin)
	// line, _ := reader.ReadString('\n')

	// fmt.Printf("%#v", line)
	// handle(line)

	list, err := filelist.Make("/Users/sergey/unisync")
	fmt.Println(err)

	if err == nil {
		for _, file := range list {
			fmt.Printf("%+v\n", file)
		}
	}

}

// func handle(line string) {
// 	words := strings.Fields(line)
// 	cmd := strings.ToUpper(words[0])

// 	switch cmd {
// 	case "REQFILELIST":
// 		return handleREQFILELIST()
// 	}

// 	fmt.Printf("%#v", cmd)
// }

// func handleREQFILELIST() error {

// }
