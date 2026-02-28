# go-reloaded — Code Breakdown

A full line-by-line explanation of how the text auto-correction tool works.

---

## Table of Contents

1. [Package & Imports](#1-package--imports)
2. [main() — The Entry Point](#2-main--the-entry-point)
3. [processFlags() — Handling Tags like (up), (hex), (cap, 3)](#3-processflags--handling-tags)
4. [processPunctuation() — Fixing Spacing Around Punctuation & Quotes](#4-processpunctuation--fixing-punctuation)
5. [processVowels() — Fixing "a" → "an"](#5-processvowels--fixing-a--an)
6. [How The 3 Passes Connect](#6-how-the-3-passes-connect)

---

## 1. Package & Imports

```go
package main
```
Every Go program that you want to **run directly** must be in `package main`. This tells Go this is an executable, not a library.

```go
import (
    "fmt"
    "os"
    "strconv"
    "strings"
)
```

| Package | Why it's used |
|---------|--------------|
| `fmt` | Printing error messages and converting numbers to strings with `fmt.Sprint()` |
| `os` | Reading/writing files (`os.ReadFile`, `os.WriteFile`) and reading command-line arguments (`os.Args`) |
| `strconv` | Converting between strings and numbers — `strconv.Atoi()` (string → int) and `strconv.ParseInt()` (string → int64 with a base like hex/binary) |
| `strings` | All string operations: splitting, trimming, replacing, case conversion |

---

## 2. `main()` — The Entry Point

This is the first function Go runs. Its job is simple: **read input → process → write output**.

```go
func main() {
```
Declares the main function. Go always starts execution here.

---

```go
    if len(os.Args) != 3 {
        return
    }
```
`os.Args` is a slice of all the command-line arguments. `os.Args[0]` is always the program name itself.

So when you run:
```
go run . sample.txt result.txt
```
- `os.Args[0]` = the program
- `os.Args[1]` = `"sample.txt"`
- `os.Args[2]` = `"result.txt"`

That's a length of **3**. If the user forgets an argument, the program quietly exits with `return` instead of crashing.

---

```go
    inputFile := os.Args[1]
    outputFile := os.Args[2]
```
Stores the two filenames into named variables so the rest of the code is readable.

---

```go
    data, err := os.ReadFile(inputFile)
```
Reads the **entire contents** of the input file into `data` (which is a `[]byte` — a slice of raw bytes). The `:=` means "declare and assign at the same time". `err` will be `nil` if successful, or contain an error description if something went wrong.

```go
    if err != nil {
        fmt.Println("Error reading file:", err)
        return
    }
```
If the file doesn't exist or can't be opened, print the error and stop. `nil` in Go means "no error / nothing".

---

```go
    text := string(data)
```
Converts the raw bytes (`[]byte`) into a regular Go `string` so we can work with it using `strings` functions.

---

```go
    text = processFlags(text)
    text = processPunctuation(text)
    text = processVowels(text)
```
Runs the text through **3 sequential passes**. Each function takes the current text, transforms it, and returns the updated version which is immediately fed into the next function. Order matters — flags are resolved first, then punctuation spacing is cleaned up, then articles are fixed.

---

```go
    if len(text) > 0 && text[len(text)-1] != '\n' {
        text += "\n"
    }
```
Ensures the output file always ends with a newline character. `text[len(text)-1]` accesses the **last character** of the string. `'\n'` is Go's way of writing the newline character. This is standard Unix file formatting.

---

```go
    err = os.WriteFile(outputFile, []byte(text), 0644)
```
Writes the final text to the output file.
- `[]byte(text)` converts the string back to bytes (that's what `WriteFile` expects)
- `0644` is the **file permission** in octal notation — it means the owner can read/write, others can only read (standard for text files)

```go
    if err != nil {
        fmt.Println("Error writing file:", err)
    }
```
If writing fails (e.g. no disk space, no permission), print the error.

---

## 3. `processFlags()` — Handling Tags

This pass finds and processes all transformation tags like `(hex)`, `(up)`, `(cap, 3)` and applies them to the words that come before them.

```go
func processFlags(text string) string {
```
Takes the full text as a string, returns the modified string.

---

```go
    words := strings.Fields(text)
```
`strings.Fields()` splits the text by **any whitespace** (spaces, tabs, newlines) and returns a slice of individual words. Extra spaces are automatically ignored.

Example: `"hello   world"` → `["hello", "world"]`

---

```go
    for i := 0; i < len(words); i++ {
        word := words[i]
```
Loops through every word by index. We use an index (`i`) instead of a range loop because we need to **modify and remove items** from the slice mid-loop, and adjust `i` accordingly.

---

```go
        if strings.HasPrefix(word, "(") {
```
A tag always starts with `(`. This filters out all normal words efficiently — we only do extra work when we see a potential tag.

---

```go
            op := ""
            count := 1
            isComplex := false
```
Sets up three variables before we figure out what kind of tag this is:
- `op` — the operation name (e.g. `"hex"`, `"up"`)
- `count` — how many words to affect (defaults to `1`)
- `isComplex` — tracks if this is a two-token tag like `(cap, 6)` which needs TWO words removed instead of one

---

### Detecting a Simple Tag: `(hex)`, `(up)`, etc.

```go
            if strings.HasSuffix(word, ")") {
                op = strings.Trim(word, "()")
            }
```
If the word both starts with `(` AND ends with `)`, it's a simple self-contained tag. `strings.Trim()` strips the `(` and `)` from both ends, leaving just the operation name.

Example: `"(up)"` → `op = "up"`

---

### Detecting a Complex Tag: `(cap, 6)`

In the words slice, `(cap, 6)` gets split into **two separate tokens**: `"(cap,"` and `"6)"`.

```go
            if strings.HasSuffix(word, ",") && i+1 <= len(words) && strings.HasSuffix(words[i+1], ")") {
```
This checks all three conditions at once:
1. The current word ends with `,` — meaning it's the first half of a complex tag
2. There IS a next word (`i+1 <= len(words)`)
3. The next word ends with `)` — confirming it's the second half

```go
                op = strings.Trim(word, "(,")
```
Strips `(` from the front and `,` from the end to extract the operation name.
Example: `"(cap,"` → `op = "cap"`

```go
                numStr := strings.Trim(words[i+1], "()")
```
Strips the `)` from the next word to get just the number string.
Example: `"6)"` → `numStr = "6"`

```go
                val, err := strconv.Atoi(numStr)
                if err == nil {
                    count = val
                    isComplex = true
                }
```
`strconv.Atoi()` converts the number string to an integer. If it succeeds (no error), we set `count` to that number and mark the tag as complex.

---

### Applying the Transformation

```go
            if op == "hex" || op == "bin" || op == "low" || op == "up" || op == "cap" {
```
Only proceed if `op` is one of the five valid operations. This guards against false positives like a word `"(hello)"` which would have passed the `HasPrefix` check.

```go
                for j := 1; j <= count && i-j >= 0; j++ {
                    target := words[i-j]
```
Loops **backwards** from the tag, `count` number of times. `j=1` means "one word before the tag", `j=2` means "two words before", etc. The `i-j >= 0` guard prevents going out of bounds at the start of the slice.

```go
                    switch op {
                    case "hex":
                        val, _ := strconv.ParseInt(target, 16, 64)
                        words[i-j] = fmt.Sprint(val)
```
`strconv.ParseInt(target, 16, 64)` converts a hexadecimal string to a base-10 `int64`. The `_` discards the error. `fmt.Sprint(val)` converts the number back to a string.
Example: `"1E"` → `30` → `"30"`

```go
                    case "bin":
                        val, _ := strconv.ParseInt(target, 2, 64)
                        words[i-j] = fmt.Sprint(val)
```
Same idea but base `2` (binary).
Example: `"1010"` → `10` → `"10"`

```go
                    case "up":
                        words[i-j] = strings.ToUpper(target)
                    case "low":
                        words[i-j] = strings.ToLower(target)
                    case "cap":
                        words[i-j] = strings.Title(strings.ToLower(target))
```
Straightforward string case transformations. `strings.Title()` capitalizes the first letter — we first `ToLower` the whole word so `"HELLO"` becomes `"Hello"` rather than `"HELLO"`.

---

### Removing the Tag from the Slice

After applying the transformation, the tag token(s) need to be deleted from the `words` slice.

```go
                if isComplex {
                    words = append(words[:i], words[i+2:]...)
                } else {
                    words = append(words[:i], words[i+1:]...)
                }
```
`append(words[:i], words[i+1:]...)` is the standard Go idiom for **removing an element at index `i`**. It takes everything before index `i` and everything after it, then joins them.

- Simple tag `(up)` → removes 1 token (at `i`)
- Complex tag `(cap,` + `6)` → removes 2 tokens (at `i` and `i+1`), so we skip ahead by `i+2`

```go
                i--
```
After removing a word, the slice is shorter and the next word has shifted into position `i`. Decrementing `i` by 1 means the next iteration of `i++` brings us back to the same index, so we don't skip the word that just moved into the tag's old position.

---

```go
    return strings.Join(words, " ")
```
Reassembles all the words back into a single string with single spaces between them.

---

## 4. `processPunctuation()` — Fixing Punctuation

This pass walks through the text **one character at a time** and rebuilds it correctly — removing spaces before punctuation and adding a space after it where needed. This approach is more efficient than the previous version because it never splits the text into a slice at all, it just reads and writes characters directly.

```go
func processPunctuation(text string) string {
    var result strings.Builder
```
`strings.Builder` is Go's most efficient way to build a string piece by piece. Instead of creating a new string every time you add a character (which is slow and wastes memory), `Builder` keeps an internal buffer and only produces the final string when you call `.String()` at the end.

---

```go
    for i := 0; i < len(text); i++ {
```
Loops through every single character in the text by its byte index. `len(text)` returns the number of bytes in the string.

---

### Rule 1: Skip spaces that appear before punctuation

```go
        if text[i] == ' ' && i+1 < len(text) && strings.ContainsRune(".,!?:;", rune(text[i+1])) {
            continue
        }
```
This is a **lookahead** — before writing the current character, we peek at the next one. Three conditions are all checked together:

1. `text[i] == ' '` — the current character is a space
2. `i+1 < len(text)` — there IS a next character (guards against going out of bounds)
3. `strings.ContainsRune(".,!?:;", rune(text[i+1]))` — the next character is punctuation

If all three are true, the space is a bad space that sits between a word and its punctuation. `continue` skips it entirely — it never gets written to `result`.

Example: `"hello ,"` → the space before `,` is skipped → `"hello,"`

---

```go
        result.WriteByte(text[i])
```
If the character was not skipped above, it gets written into the result buffer. `WriteByte()` adds a single byte (character) to the builder.

---

### Rule 2: Insert a space after punctuation if one is missing

```go
        if strings.ContainsRune(".,!?:;", rune(text[i])) && i+1 < len(text) && !strings.ContainsRune(".,!?:; ", rune(text[i+1])) {
            result.WriteByte(' ')
        }
```
After writing a punctuation character, we immediately check if a space needs to be added after it. Again three conditions:

1. `strings.ContainsRune(".,!?:;", rune(text[i]))` — the character we just wrote is punctuation
2. `i+1 < len(text)` — there is a next character
3. `!strings.ContainsRune(".,!?:; ", rune(text[i+1]))` — the next character is **not** another punctuation mark AND **not** already a space

The `!` (not) on condition 3 is what handles groups like `...` or `!?` — if the next character is also punctuation, we don't insert a space between them. We only add a space when the next character is a real word character.

Example walkthrough with `"hello,world"`:
- Writes `h`, `e`, `l`, `l`, `o` normally
- Writes `,` → next character is `w`, which is not punctuation or space → inserts `' '`
- Writes `w`, `o`, `r`, `l`, `d` normally
- Result: `"hello, world"`

Example with `"hello..."`:
- Writes `.` → next is `.` → condition 3 is false → no space inserted
- Writes `.` → next is `.` → no space inserted
- Writes `.` → nothing after → condition 2 is false → no space inserted
- Result: `"hello..."` ✓

---

```go
    return result.String()
```
`.String()` finalizes the builder and returns everything that was written to it as a single string.

---

## 5. `processVowels()` — Fixing "a" → "an"

```go
func processVowels(text string) string {
    words := strings.Fields(text)
```
Splits the text into words again for processing.

---

```go
    for i := 0; i < len(words)-1; i++ {
```
Loops through all words **except the last one** (`len(words)-1`) because we always need to look ahead at `words[i+1]`. If we went to the last word, `words[i+1]` would be out of bounds.

---

```go
        if (words[i] == "a" || words[i] == "A") && strings.ContainsRune("aeiouAEIOUhH", rune(words[i+1][0])) {
            words[i] += "n"
        }
```
Two conditions must both be true:

1. **`words[i] == "a" || words[i] == "A"`** — the current word is exactly the article "a" or "A" (case-sensitive exact match, so "and" or "at" are not affected)

2. **`strings.ContainsRune("aeiouAEIOUhH", rune(words[i+1][0]))`**
   - `words[i+1][0]` → gets the **first byte** of the next word
   - `rune(...)` → converts it to a Unicode character (a `rune`)
   - `strings.ContainsRune("aeiouAEIOUhH", ...)` → checks if that character is in our vowel+h string

If both are true, we append `"n"` to the article: `"a"` → `"an"`, `"A"` → `"An"`.

---

```go
    return strings.Join(words, " ")
```
Joins everything back into a single string.

---

## 6. How The 3 Passes Connect

```
Input text
    │
    ▼
processFlags()
    Resolves all (hex), (bin), (up), (low), (cap), (cap, N) tags
    Removes the tags from the text
    │
    ▼
processPunctuation()
    Fixes spacing around . , ! ? : ;
    Fixes single quote spacing
    │
    ▼
processVowels()
    Converts "a" → "an" before vowels and 'h'
    │
    ▼
Output text written to file
```

**Why this order?**

- Flags must go first because `(cap, 2)` works on raw words — if punctuation was already attached, `"times,"` and `"times"` would be treated differently.
- Punctuation goes second because after flags are resolved, word spacing is clean and predictable.
- Vowels go last so that if a flag like `(low)` turned `"AN"` into `"an"`, the article check still works correctly on the final casing.
