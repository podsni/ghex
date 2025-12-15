package ui

import (
	"fmt"
	"sync"
	"time"
)

// Spinner represents a loading spinner
type Spinner struct {
	message string
	frames  []string
	current int
	done    chan bool
	mu      sync.Mutex
	running bool
}

// NewSpinner creates a new spinner with a message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:    make(chan bool),
	}
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				frame := s.frames[s.current]
				s.current = (s.current + 1) % len(s.frames)
				s.mu.Unlock()

				// Clear line and print spinner
				fmt.Printf("\r%s %s %s", PrimaryStyle.Render(frame), TextStyle.Render(s.message), "")
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.done <- true

	// Clear the spinner line
	fmt.Print("\r\033[K")
}

// StopWithMessage stops the spinner and displays a final message
func (s *Spinner) StopWithMessage(message string) {
	s.Stop()
	fmt.Println(message)
}

// StopWithSuccess stops the spinner and displays a success message
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	ShowSuccess(message)
}

// StopWithError stops the spinner and displays an error message
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	ShowError(message)
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// WithSpinner executes a function while showing a spinner
func WithSpinner(message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()
	err := fn()
	if err != nil {
		spinner.StopWithError(err.Error())
	} else {
		spinner.Stop()
	}
	return err
}
