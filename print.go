package main

import (
	"fmt"

	"github.com/fatih/color"
)

func PrintFAIL(text string) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("> FAIL\t")
	fmt.Printf("%s\n", text)

}

func PrintPASS(text string) {
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)
	boldGreen.Printf("> PASS\t")
	fmt.Printf("%s\n", text)
}

func PrintWARN(text string) {
	red := color.New(color.FgYellow)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("> WARN\t")
	fmt.Printf("%s\n", text)
}

func PrintStep(text string) {
	stepCount++
	bold := color.New(color.Bold)
	bold.Printf("[Step%d]\t", stepCount)
	fmt.Printf("%s\n", text)
}
