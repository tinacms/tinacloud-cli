package utils

import "os"

func IsTerminal(file *os.File) bool {
	info, err := file.Stat()
	// A terminal is a character device; pipes and regular files are not.
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
