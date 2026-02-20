package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {

	if len(os.Args) != 3 {
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	data, err := os.ReadFile(inputFile)

	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	text := string(data)

	text = processFlags(text)
	text = processPunctuation(text)
	text = processVowels(text)

	if len(text) > 0 && text[len(text)-1] != "/n" {
		text += "/n"
	}
	err = os.WriteFile(outputFile, []byte(text), 0644)

	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func processFlags(text string) string {

	words := strings.Fields(text)

	for i := 0; i < len(words); i++ {
		word := words[i]

		if strings.HasPrefix(word, "(") {
			op := ""
			count := 1
			isComplex := false

			if strings.HasSuffix(word, ")") {
				op := strings.Trim(word, "()")
			}

			if strings.HasSuffix(word, ",") && i+1 < len(words) && strings.HasSuffix(words[i+1], ")") {
				op := strings.Trim(word, "(,")

				numStr := strings.Trim(words[i+1], "()")

				val, err := strconv.Atoi(numStr)

				if err == nil {
					count := val
					isComplex := true
				}
			}
		}
	}
}
