package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	ScrambledString string
	TerminatingChar string
	WordFile        os.File
	Limit           int
	Min             int
	Max             int
	SortMethod      string
	SortReverse     bool
	JsonOutput      bool
	Filter          string
	DebugLevel      int
}

type Letter struct {
	Self     string             `json:"self"`
	Children map[string]*Letter `json:"children,omitempty"`
}

// Logger construct and functions are defined below
type Logger struct {
	*log.Logger
	level int
}

func newLogger(prefix string, level int) *Logger {
	return &Logger{
		Logger: log.New(os.Stderr, prefix, 0),
		level:  level,
	}
}

func (l *Logger) Log(level int, v ...interface{}) {
	colour := "\u001B[0;m"
	switch {
	case level >= 10 && level < 20:
		colour = "\u001B[5;90;m"
	case level >= 20 && level < 30:
		colour = "\u001B[0;m"
	case level >= 30 && level < 40:
		colour = "\u001B[5;32;m"
	case level >= 40 && level < 50:
		colour = "\u001B[5;93;m"
	case level >= 50 && level < 60:
		colour = "\u001B[5;41;97;m"
	}
	endColour := "\u001B[0m"

	if level >= l.level {
		// Add colour to first element if there is a firs element
		if len(v) > 0 {
			str, _ := v[0].(string)
			v[0] = colour + str + endColour
		}
		l.Logger.Println(v...)
	}
	return
}

func (l *Logger) VerboseDebug(message string, v ...interface{}) {
	l.Log(10, fmt.Sprintf(message, v...))
}

func (l *Logger) Debug(message string, v ...interface{}) {
	l.Log(11, fmt.Sprintf(message, v...))
}

func (l *Logger) Info(message string, v ...interface{}) {
	l.Log(20, fmt.Sprintf(message, v...))
}

func (l *Logger) Warning(message string, v ...interface{}) {
	l.Log(30, fmt.Sprintf(message, v...))
}

func (l *Logger) Error(message string, v ...interface{}) {
	l.Log(40, fmt.Sprintf(message, v...))
}

func (l *Logger) Critical(message string, v ...interface{}) {
	l.Log(50, fmt.Sprintf(message, v...))
	os.Exit(254)
}

// Main
func main() {
	// Let's track how long tasks take
	startTime := time.Now()

	// Loaf options and start logging
	options := getOptions()
	logger := newLogger("", options.DebugLevel)

	// Debug the options
	v := reflect.ValueOf(options)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		fieldValue := field.Interface()
		logger.Debug("%s: %v", fieldName, fieldValue)
	}
	logger.Debug("File: %s", options.WordFile.Name())
	logger.Debug("%-60s %8.1fms\n", "Loaded options:", float64(time.Since(startTime).Milliseconds()))

	logger.Debug("Debug ON")
	logger.Info("Starting")

	// Read the word file content
	content, err := os.ReadFile(options.WordFile.Name())
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("%-15s %-45s%8.1fms\n", "Loaded file:", options.WordFile.Name(), float64(time.Since(startTime).Milliseconds()))

	// Split the content into lines
	lines := splitByAnyHiddenCharacters(string(content))

	// Create a root for the dictionary tree.  Similar to https://github.com/thedzy/boggle_solver except we generate it live
	root := &Letter{
		Children: make(map[string]*Letter),
	}

	// Load the words into the tree
	logger.Info("Loading %s words into dictionary", prettyFormatInt(len(lines)))
	for _, line := range lines {
		addWord(root, line)
	}

	logger.Debug("%-60s %8.1fms\n", "Built word tree:", float64(time.Since(startTime).Milliseconds()))

	// Debugging the tree
	jsonBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		logger.Error("Error marshaling JSON:", err)
		return
	}
	logger.VerboseDebug(string(jsonBytes))

	if options.Max == 0 {
		options.Max = len(options.ScrambledString)
	}
	logger.Info("Finding words of %d to %d length", options.Min, options.Max)
	matches := getVariations(root, strings.ToLower(options.ScrambledString), options.Min, options.Max)
	logger.Debug("%-60s %8.1fms\n", "Found matches:", float64(time.Since(startTime).Milliseconds()))
	logger.Info("Found %d words", len(matches))

	// Sort/filter the matches
	if len(options.Filter) != 0 {
		var filteredMatches []string

		re := regexp.MustCompile(options.Filter)
		for _, match := range matches {
			if re.MatchString(match) {
				filteredMatches = append(filteredMatches, match)
			} else {
				logger.VerboseDebug(match)
			}
		}
		matches = filteredMatches
		logger.Debug("Applied filter %s", options.Filter)
	}

	if options.SortMethod == "a" || options.SortMethod == "alpha" {
		sort.Strings(matches)
		logger.Debug("Sorted alphabetically")
	}

	if options.SortMethod == "l" || options.SortMethod == "len" {
		sort.Slice(matches, func(i, j int) bool {
			return len(matches[i]) < len(matches[j])
		})
		logger.Debug("Sorted by length")
	}
	if options.SortReverse {
		// Reverse the order of elements in the list
		sort.SliceStable(matches, func(i, j int) bool {
			return i > j
		})
		logger.Debug("Reversed sorting")
	}
	if options.Limit > 0 && options.Limit < len(matches) {
		matches = matches[0:options.Limit]
		logger.Debug("Trimmed list to %d", options.Limit)
	}
	logger.Debug("%-60s %8.1fms\n", "Sorted and filtered:", float64(time.Since(startTime).Milliseconds()))
	logger.Info("Displaying %d words", len(matches))
	logger.Info("----")

	// Print results
	if options.JsonOutput {
		if len(matches) == 0 {
			fmt.Println("[]")
		} else {
			jsonData, err := json.Marshal(matches)
			if err != nil {
				logger.Error("Error:", err)
				return
			}
			fmt.Println(string(jsonData))
		}
	} else {
		for _, match := range matches {
			//logger.Info(match)
			fmt.Println(match)
		}
	}

	logger.Info("----")
	logger.Info("Done")

	logger.Debug("%-60s %8.1fms", "Printed words: ", float64(time.Since(startTime).Milliseconds()))

}

func addWord(parent *Letter, word string) {
	// addWord: appends another word to the dictionary tree.
	//
	// It follows the tree until a new branch is needed and branches from there
	// the parent node is the parent from the last branch or the root

	// Convert to lowercase
	word = strings.ToLower(word) + "\n"

	for _, char := range word {
		letter := string(char)

		if _, exists := parent.Children[letter]; !exists {
			parent.Children[letter] = &Letter{
				Self:     letter,
				Children: make(map[string]*Letter),
			}
		}
		parent = parent.Children[letter]
	}
}

func searchWord(node *Letter, word string) bool {
	// searchWord: recursively follows the tree until it can't find a branch.
	//
	// This can be used to find partial or complete branches.
	// Have a terminator to the end of the word differentiates between partial and full

	// If we have exhausted all letters than the search is complete, and we have not reached the point where there is no branches left
	if len(word) == 0 {
		return true
	}

	// If the next branch doesn't exist we fail, otherwise we follow the next
	letter := string(word[0])
	nextNode, exists := node.Children[letter]
	if !exists {
		return false
	}

	return searchWord(nextNode, word[1:])
}

func getVariations(node *Letter, letters string, min int, max int) []string {
	// getVariations: acts like a container to hold the results from the searches.
	//
	// It follows the tree until a new branch is needed and branches from there
	// the parent node is the parent from the last branch or the root

	var combinations []string
	used := make([]bool, len(letters))
	builder := strings.Builder{}

	searchVariant(node, letters, used, &builder, &combinations, min, max)

	return combinations
}

func searchVariant(node *Letter, letters string, used []bool, builder *strings.Builder, combinations *[]string, min int, max int) {
	// searchVariant: is where I got the idea for this, I was looking to create something like th boggle solver in go
	//
	// It takes the first letters and then starts a process for each subsequent letter until we hit the end

	// If the word is in range and if it can be matched fully
	if builder.Len() >= min && builder.Len() <= max && builder.Len() != 0 {
		if searchWord(node, builder.String()+"\n") {
			if !inList(*combinations, builder.String()) {
				*combinations = append(*combinations, builder.String())
			}
		}
		//return
	}

	// Start looping through he first letters
	for i := 0; i < len(letters); i++ {
		if used[i] {
			continue
		}

		used[i] = true
		builder.WriteByte(letters[i])

		// Keep checking that the word can possibly match
		// Speeds up search by not following words that will never be words
		if searchWord(node, builder.String()) {
			//fmt.Println(builder.String())
			searchVariant(node, letters, used, builder, combinations, min, max)
		} else {
			//fmt.Println(builder.String())
		}

		used[i] = false
		newStr := builder.String()[:builder.Len()-1]
		builder.Reset()
		builder.WriteString(newStr)
	}
}

func inList(list []string, str string) bool {
	// inList: is the word already in a list/
	//
	// I couldn't find a way to do this natively like in Python, it checks that a word is not already in a list

	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

func splitByAnyHiddenCharacters(input string) []string {
	//splitByAnyHiddenCharacters: Use regex to split the string by control character
	//
	// Now it shouldn't matter what platform you are on
	//https://pkg.go.dev/regexp/syntax

	regex := regexp.MustCompile(`[[:cntrl:]]+`)
	return regex.Split(input, -1)
}

func prettyFormatInt(number int) string {
	// prettyFormatInt: I wanted to print a number with comma separators

	digits := strconv.Itoa(number)
	var groups []string
	for x := len(digits); x >= 1; x -= 3 {
		if x >= 3 {
			groups = append([]string{digits[x-3 : x]}, groups...)
		} else {
			groups = append([]string{digits[0:x]}, groups...)
		}
	}
	return strings.Join(groups, ",")
}

func getOptions() Options {
	// getOptions: get the program options.  If anyone has a better way, please comment
	//
	// I like the way python handles options and this is the best I could find without essentially recreating this

	// Get path for options
	dir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	// Check for stdin
	var lettersDefault string
	var lettersRequired = true

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Get stdin and let to the default in options, setting option to optional
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lettersDefault = scanner.Text()

			// Remove control characters
			re := regexp.MustCompile("`[[:cntrl:]]+")
			lettersDefault = re.ReplaceAllString(scanner.Text(), "")
			lettersRequired = false
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading stdin", err)
			lettersRequired = true
		}
	}

	// Create parser for parsing the options
	parser := argparse.NewParser("unscramble", "take some letters and arrange them in different orders and find the words")
	var version = parser.Flag(
		"v", "version",
		&argparse.Options{
			Help: "Current version"})

	// Input options
	var scrambledString = parser.String("l", "letters",
		&argparse.Options{
			Help: "Letters ",
			Validate: func(args []string) error {
				for _, arg := range args {
					re := regexp.MustCompile("[[:cntrl:]]")
					if re.MatchString(arg) {
						return fmt.Errorf("number must not contain contol characters")
					}
				}
				return nil
			},
			Default:  lettersDefault,
			Required: lettersRequired})

	var terminatingChar = parser.String("t", "terminator",
		&argparse.Options{
			Help:    "Any existing terminating characters on word files lines, Ex. \\r ",
			Default: nil})

	var wordFile = parser.File("f", "file", os.O_RDONLY, 0600,
		&argparse.Options{
			Help:    "Words file ",
			Default: filepath.Join(dir, "collins_scrabble_words_2019.txt")})

	// Output options
	var sortMethod = parser.Selector("s", "sort", []string{"a", "alpha", "l", "len"},
		&argparse.Options{
			Help:    "Sorting method ",
			Default: nil})

	var sortReverse = parser.Flag("r", "sort-reverse",
		&argparse.Options{
			Help:    "Sort reversed",
			Default: false})

	var limit = parser.Int("", "limit",
		&argparse.Options{
			Help: "Limit to x results"})

	var min = parser.Int("", "min",
		&argparse.Options{
			Help:    "Length of the smallest word",
			Default: 0})

	var max = parser.Int("", "max",
		&argparse.Options{
			Help:    "Length of the largest word",
			Default: nil})

	var jsonOutput = parser.Flag("j", "json",
		&argparse.Options{
			Help:    "Json output",
			Default: false})

	var filter = parser.String("", "filter",
		&argparse.Options{
			Help:    "Filter output with regex, Ex ^a.*[ety]$",
			Default: nil})

	var debugLevel = parser.Int("", "log-level",
		&argparse.Options{
			Help: "Set the logging level",
			Validate: func(args []string) error {
				for _, arg := range args {
					number, err := strconv.Atoi(arg)
					if err != nil {
						return fmt.Errorf("invalid number: %s", arg)
					}
					if number < 10 || number > 60 {
						return fmt.Errorf("number must be between 10 and 60")
					}
				}
				return nil
			},
			Default: 20})

	// Parsing options
	err = parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}

	// Display version number
	if *version {
		print("1.5", "\n")
		os.Exit(0)
	}

	return Options{
		ScrambledString: *scrambledString,
		TerminatingChar: *terminatingChar,
		WordFile:        *wordFile,
		Limit:           *limit,
		Min:             *min,
		Max:             *max,
		SortMethod:      *sortMethod,
		SortReverse:     *sortReverse,
		JsonOutput:      *jsonOutput,
		Filter:          *filter,
		DebugLevel:      *debugLevel,
	}
}
