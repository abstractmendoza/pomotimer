package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
  
  "github.com/eiannone/keyboard"
)

const (
	defaultWorkPeriod      = 25 * time.Minute
	defaultBreakPeriod     = 1 * time.Minute
	defaultLongBreakPeriod = 15 * time.Minute
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pomo <action>")
		fmt.Println("Actions: work, break, lbreak")
		return
	}

	action := strings.ToLower(os.Args[1])

	switch action {
	case "work":
		timer(defaultWorkPeriod)
	case "break":
		timer(defaultBreakPeriod)
	case "lbreak":
		timer(defaultLongBreakPeriod)
	default:
		fmt.Printf("%s is not an available action\n", action)
		fmt.Println("Actions: work, break, lbreak")
	}
}

func timer(duration time.Duration) {
	remaining := duration
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM)

	paused := false

  keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	go func() {
    for {
      event := <-keysEvents
      if event.Err != nil {
        panic(event.Err)
      }
      if event.Key == keyboard.KeyCtrlC {
        killProgram()
      }
      if event.Rune == 'p' {
        paused = !paused
      }
    }
	}()

	for remaining > 0 {
		select {
		case <-exitSignal:
			clearLine()
			fmt.Print("Timer cancelled")
			return
		case <-ticker.C:
			if !paused {
				remaining -= time.Second
				clearLine()
				fmt.Printf("Remaining %02d:%02d", int(remaining.Minutes()), int(remaining.Seconds())%60)
			}
		}
	}

	clearLine()
	fmt.Println("\nFinished")
}

func killProgram() {
  pid := os.Getpid()
  err := syscall.Kill(pid, syscall.SIGINT)

  if err != nil {
    panic(err)
  }
}

func clearLine() {
	fmt.Print("\033[1G\033[2K") // ANSI scape code: Move to the start of line and delete the current content
}
