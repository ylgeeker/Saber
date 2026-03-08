/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// crlfWriter wraps w and converts \n to \r\n so that in raw terminal mode each newline
// moves the cursor to column 0, fixing help/command output formatting.
type crlfWriter struct{ w io.Writer }

func (c *crlfWriter) Write(p []byte) (n int, err error) {
	q := bytes.ReplaceAll(p, []byte("\n"), []byte("\r\n"))
	_, err = c.w.Write(q)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// stdio ties stdin and stdout into a single io.ReadWriter for use with term.NewTerminal.
type stdio struct{}

func (*stdio) Read(p []byte) (n int, err error)  { return os.Stdin.Read(p) }
func (*stdio) Write(p []byte) (n int, err error) { return os.Stdout.Write(p) }

func setupGracefulShutdown(svr *Service) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigC
		svr.Close()
		os.Exit(0)
	}()
}

// runInteractive uses golang.org/x/term.Terminal for line editing, echo, and Ctrl+L clear.
// See: https://pkg.go.dev/golang.org/x/term#Terminal
func runInteractive(root *cobra.Command) error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return runInteractiveStdin(root)
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, oldState)

	t := term.NewTerminal(&stdio{}, "SABER> ")
	for {
		line, err := t.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			if err == term.ErrPasteIndicator {
				continue
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if handleInteractiveCommand(root, line, func() { clearScreenTerminal(t) }) {
			return nil
		}
	}
	return nil
}

// clearScreenTerminal sends VT100 clear-screen and home cursor to the terminal.
func clearScreenTerminal(t *term.Terminal) {
	t.Write([]byte("\x1b[2J\x1b[H"))
}

// runInteractiveStdin is used when stdin is not a TTY (e.g. pipe); no raw mode, no Ctrl+L.
func runInteractiveStdin(root *cobra.Command) error {
	buf := make([]byte, 0, 256)
	for {
		os.Stdout.WriteString("SABER> ")
		buf = buf[:0]
		for {
			var b [1]byte
			n, err := os.Stdin.Read(b[:])
			if err != nil {
				if err == io.EOF && len(buf) > 0 {
					break
				}
				return err
			}
			if n == 0 {
				continue
			}

			c := b[0]
			if c == '\n' || c == '\r' {
				break
			}

			buf = append(buf, c)
			os.Stdout.Write(b[:])
		}

		line := strings.TrimSpace(string(buf))
		if line == "" {
			continue
		}

		if handleInteractiveCommand(root, line, func() { os.Stdout.WriteString("\x1b[2J\x1b[H") }) {
			return nil
		}
	}
}

// handleInteractiveCommand handles one interactive line. Returns true if the session should exit (exit/quit).
func handleInteractiveCommand(root *cobra.Command, line string, clearFn func()) bool {
	lower := strings.ToLower(line)
	if lower == "exit" || lower == "quit" {
		return true
	}
	if lower == "clear" {
		clearFn()
		return false
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}
	subName := strings.ToLower(parts[0])
	for _, c := range root.Commands() {
		if c.Name() == subName {
			root.SetArgs(parts)
			// In raw terminal mode, \n does not move cursor to column 0; wrap stdout/stderr
			// so that \n becomes \r\n and help/output format correctly.
			root.SetOut(&crlfWriter{w: os.Stdout})
			root.SetErr(&crlfWriter{w: os.Stderr})
			err := root.ExecuteContext(context.Background())
			root.SetOut(nil)
			root.SetErr(nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\r\n", err)
			}
			return false
		}
	}
	fmt.Fprintf(os.Stderr, "Unknown command %q. Try 'help' or 'help list'.\r\n", parts[0])
	return false
}
