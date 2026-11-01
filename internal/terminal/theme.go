// Package terminal applies Nord accents only to interactive, color-enabled output.
package terminal

import (
	"io"
	"os"
)

// Interactive reports whether w is a terminal (not a pipe, file or log).
func Interactive(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func colorEnabled(w io.Writer) bool {
	_, noColor := os.LookupEnv("NO_COLOR")
	return Interactive(w) && !noColor && os.Getenv("TERM") != "dumb"
}

// Frost writes text in Nord Frost (nord8).
func Frost(w io.Writer, text string) { paint(w, "136;192;208", text) }

// Snow writes text in Nord Snow Storm (nord6), for values meant to stand out.
func Snow(w io.Writer, text string) { paint(w, "236;239;244", text) }

// Dim writes secondary text in Nord Snow Storm nord4.
func Dim(w io.Writer, text string) { paint(w, "216;222;233", text) }

func paint(w io.Writer, rgb, text string) {
	if colorEnabled(w) {
		_, _ = io.WriteString(w, "\x1b[38;2;"+rgb+"m"+text+"\x1b[0m")
		return
	}
	_, _ = io.WriteString(w, text)
}
