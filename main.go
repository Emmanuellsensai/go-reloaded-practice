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

	if len(text) > 0 && text[len(text)-1] != '\n' {
		text += "\n"
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

				op = strings.Trim(word, "()")

			}

			if strings.HasSuffix(word, ",") && i+1 <= len(words) && strings.HasSuffix(words[i+1], ")") {

				op = strings.Trim(word, "(,")

				numStr := strings.Trim(words[i+1], "()")

				val, err := strconv.Atoi(numStr)

				if err == nil {

					count = val
					isComplex = true

				}
			}

			if op == "hex" || op == "bin" || op == "low" || op == "up" || op == "cap" {

				for j := 1; j <= count && i-j >= 0; j++ {

					target := words[i-j]

					switch op {
					case "hex":
						val, _ := strconv.ParseInt(target, 16, 64)
						words[i-j] = fmt.Sprint(val)

					case "bin":
						val, _ := strconv.ParseInt(target, 2, 64)
						words[i-j] = fmt.Sprint(val)

					case "up":
						words[i-j] = strings.ToUpper(target)

					case "low":
						words[i-j] = strings.ToLower(target)

					case "cap":
						words[i-j] = strings.Title(strings.ToLower(target))
					}
				}
				if isComplex {

					words = append(words[:i], words[i+2:]...)

				} else {

					words = append(words[:i], words[i+1:]...)
				}
				i--
			}
		}
	}
	return strings.Join(words, " ")
}

// PASS 2: Handle Punctuation spacing & Single Quotes
func processPunctuation(text string) string {
	punctuations := []string{".", ",", "!", "?", ":", ";"}

	// Step 1: Ensure space around all punctuation so Fields() isolates them
	for _, p := range punctuations {
		text = strings.ReplaceAll(text, p, " "+p+" ")
	}

	// Step 2: Attach each punctuation mark to the word before it
	words := strings.Fields(text)
	var result []string

	for _, word := range words {
		isPunc := strings.ContainsAny(word, ".!?,;:")
		if isPunc && len(result) > 0 {
			result[len(result)-1] += word // glue to previous word
		} else {
			result = append(result, word)
		}
	}

	text = strings.Join(result, " ")

	// Step 3: Fix single quotes — remove spaces inside ' ... '
	for strings.Contains(text, "' ") || strings.Contains(text, " '") {
		text = strings.ReplaceAll(text, "' ", "'")
		text = strings.ReplaceAll(text, " '", "'")
	}

	return text
}

// PASS 3: 'a' -> 'an' before vowels or 'h'
func processVowels(text string) string {
	words := strings.Fields(text)

	for i := 0; i < len(words)-1; i++ {
		if (words[i] == "a" || words[i] == "A") && strings.ContainsRune("aeiouAEIOUhH", rune(words[i+1][0])) {
			words[i] += "n"
		}
	}

	return strings.Join(words, " ")
}
