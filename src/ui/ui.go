package ui

import (
	"fmt"
	"os"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func Success(format string, a ...any) {
	fmt.Printf(ColorGreen+"✅ "+format+ColorReset+"\n", a...)
}

func Info(format string, a ...any) {
	fmt.Printf(ColorCyan+"ℹ️  "+format+ColorReset+"\n", a...)
}

func Warn(format string, a ...any) {
	fmt.Printf(ColorYellow+"⚠️  "+format+ColorReset+"\n", a...)
}

func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, ColorRed+"❌ "+format+ColorReset+"\n", a...)
}

func Heading(format string, a ...any) {
	fmt.Printf("\n"+ColorCyan+"🚀 "+format+ColorReset+"\n", a...)
}

func Finished(format string, a ...any) {
	fmt.Printf("\n"+ColorGreen+"✨ "+format+ColorReset+"\n", a...)
}
