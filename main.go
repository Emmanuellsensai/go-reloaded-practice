package main

import (
	"fmt"
	"os"
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
