package pkg

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

func ConcurencyHelp(inputchan chan string, name string) (size int) {
	currsize := len(inputchan)
	var count int
	for count < 10 {
		currsize = len(inputchan)
		str := fmt.Sprintf("%s   -> size: %d , capacity: %d \n", name, currsize, cap(inputchan))
		fmt.Println(str)
		time.Sleep(2 * time.Second)
		count++
	}
	return currsize
}

func CreateLogFile(prefix string) *os.File {
	fileName := fmt.Sprintf("%s_%s", prefix, ".log")
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to create/open log file: %v", err)
	}
	return logFile
}

// TextStyler struct manages predefined styles for text
type TextStyler struct {
	red            *color.Color
	green          *color.Color
	yellow         *color.Color
	blue           *color.Color
	boldRed        *color.Color
	underlineGreen *color.Color
	mu             sync.Mutex
}

// NewTextStyler initializes the TextStyler with predefined styles
func NewTextStyler() *TextStyler {
	return &TextStyler{
		red:            color.New(color.FgRed),
		green:          color.New(color.FgGreen),
		yellow:         color.New(color.FgYellow, color.Bold),
		blue:           color.New(color.FgBlue),
		boldRed:        color.New(color.FgRed, color.Bold),
		underlineGreen: color.New(color.FgGreen, color.Underline),
	}
}

// Red prints text in red
func (ts *TextStyler) Red(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.red.Println(text)
}

// Green prints text in green
func (ts *TextStyler) Green(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.green.Println(text)
}

func (ts *TextStyler) Yellow(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.yellow.Println(text)
}

// Blue prints text in blue
func (ts *TextStyler) Blue(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.blue.Println(text)
}

// BoldRed prints text in bold red
func (ts *TextStyler) BoldRed(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.boldRed.Println(text)
}

// UnderlineGreen prints text in underlined green
func (ts *TextStyler) UnderlineGreen(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.underlineGreen.Println(text)
}
