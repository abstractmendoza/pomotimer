package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

const (
	defaultWorkPeriod      = 25 * time.Minute
	defaultBreakPeriod     = 5 * time.Minute
	defaultLongBreakPeriod = 15 * time.Minute
)

type TimerType struct {
	label    string
	duration time.Duration
	color    *color.Color
	endSound TimerEndSound
}

type TimerStatus int

const (
	Paused TimerStatus = iota
	Running
)

type TimerEndSound int

const (
	WorkSound TimerEndSound = iota
	BreakSound
)

func main() {
	workTimer := TimerType{
		label:    "Work",
		duration: defaultWorkPeriod,
		color:    color.New(color.FgMagenta),
		endSound: WorkSound,
	}
	breakTimer := TimerType{
		label:    "Break",
		duration: defaultBreakPeriod,
		color:    color.New(color.FgGreen),
		endSound: BreakSound,
	}
	longBreakTimer := TimerType{
		label:    "Long break",
		duration: defaultLongBreakPeriod,
		color:    color.New(color.FgBlue),
		endSound: BreakSound,
	}

	drawHeader()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM)

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var currentTimer *TimerType
	remaining := 0 * time.Second
	paused := false

	for {
		select {
		case <-ticker.C:
			if remaining == 0 && currentTimer != nil {
				if currentTimer.endSound == BreakSound {
					playWavSound("break-end-sound.wav")
				} else {
					playWavSound("end-sound.wav")
				}

				clearLine()
				fmt.Printf(">>>>>> %s has finish <<<<<<", currentTimer.label)

				currentTimer = nil
			}
			if remaining > 0 {
				if paused {
					printTimer(currentTimer, Paused, remaining)
					continue
				}

				remaining -= time.Second
				printTimer(currentTimer, Running, remaining)
			}
		case <-exitSignal:
			clearLine()
			fmt.Println("The end...")
			return
		case k := <-keysEvents:
			if k.Key == keyboard.KeyCtrlC {
				killProgram()
			}

			if k.Rune == 'w' && remaining == 0 {
				currentTimer = &workTimer
				remaining = workTimer.duration
			}
			if k.Rune == 'b' && remaining == 0 {
				currentTimer = &breakTimer
				remaining = breakTimer.duration
			}
			if k.Rune == 'l' && remaining == 0 {
				currentTimer = &longBreakTimer
				remaining = longBreakTimer.duration
			}

			if k.Rune == 'p' {
				paused = !paused
			}
		}
	}
}

func playWavSound(file string) {
	soundFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	streamer, format, err := wav.Decode(soundFile)
	if err != nil {
		panic(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

func printTimer(timer *TimerType, status TimerStatus, remaining time.Duration) {
	progress := (float32(timer.duration-remaining) / float32(timer.duration))
	color := selectTimerColor(timer.color, status)

	clearLine()
	color.Print(timer.label, " ")
	color.Print(getProgressBar(progress), " ")
	if status == Paused {
		color.Print("[paused] ")
	}
	color.Printf(
		"%02d:%02d",
		int(remaining.Minutes()),
		int(remaining.Seconds())%60,
	)
}

func selectTimerColor(timerColor *color.Color, status TimerStatus) *color.Color {
	if status == Paused {
		return color.New(color.FgHiRed)
	}
	return timerColor
}

func getProgressBar(progress float32) string {
	bar := ""
	barLength := 40

	for i := 0; i < barLength; i++ {
		if progress >= float32((i+1))/float32(barLength) {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return bar
}

func drawHeader() {
	tomatoAsciiArt := `
============================================================

       """""  """""
    """     ""     """
       #####""#####       
      ##############      Pomodore timer
      ##############      [w]: Start work timer
     ################     [b]: Start break timer
     ################     [l]: Start long break timer
      #############       [p]: Pause/resume current timer
       ###########        
         ######           

============================================================
  `

	color.Red(tomatoAsciiArt)
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
