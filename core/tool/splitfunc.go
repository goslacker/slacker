package tool

import (
	"bufio"
	"bytes"
)

// SplitByEmptyLine is a SplitFunc that splits the input into tokens separated by empty lines.
func SplitByEmptyLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// If we're at EOF and have no data, return an error to stop scanning.
	if atEOF && len(data) == 0 {
		return 0, nil, bufio.ErrFinalToken
	}

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' && i+1 < len(data) && data[i+1] == '\n' {
			return i + 2, data[:i], nil
		}
	}

	if atEOF {
		return len(data), bytes.Trim(data, "\n"), bufio.ErrFinalToken
	}

	// If the next line is not empty, return the current token and advance past the newline.
	return 0, nil, nil
}
