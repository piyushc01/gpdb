package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/greenplum-db/gpdb/gpservice/constants"
)

func AskUserYesOrNo(prompt string) bool {
	fmt.Println()
	input := make(chan string, 1)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s  Yy|Nn: ", prompt)
		go func() {
			result, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Failed to read input: %v, defaulting to no\n", err)
				result = "n"
			}

			input <- strings.ToLower(strings.TrimSpace(result))
		}()

		select {
		case in := <-input:
			switch in {
			case "y":
				return true
			case "n":
				return false
			default:
				fmt.Printf("invalid input %q\n", in)
				continue
			}
		case <-time.After(constants.UserInputWaitDurtion * time.Second):
			fmt.Println("\ntimed out, defaulting to no")
			return false
		}

	}
}
