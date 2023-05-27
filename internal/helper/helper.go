package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// TruncateFileName will turncate filename for progresbar if its longer then 20 chars
func TruncateFileName(srcFilePath string) string {
	parts := strings.Split(srcFilePath, "/")
	lastTwoParts := ".." + strings.Join(parts[len(parts)-2:], "/")

	if len(lastTwoParts) > 32 {
		_, file := filepath.Split(srcFilePath)

		if len(file) > 32 {
			truncated := ".." + file[len(file)-20:]
			return truncated
		}

		return file
	}

	return lastTwoParts
}

// ProgressBar function will return *progressbar.ProgressBar with given inputs
func ProgressBar(size int64, truncatedFilePath string, onCompletionFunc func()) *progressbar.ProgressBar {
	// Customize the progress bar theme
	theme := progressbar.Theme{
		Saucer:        "\x1b[38;5;215m▖[reset][cyan]",
		SaucerPadding: " ",
	}

	progress := progressbar.NewOptions64(
		size,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(theme),
		progressbar.OptionSetDescription("Copying"+" "+"[green]"+truncatedFilePath+"[reset]"),
		progressbar.OptionSetWidth(20),
		progressbar.OptionShowBytes(true),
		progressbar.OptionOnCompletion(onCompletionFunc),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSpinnerType(14),
	)

	return progress
}

// checkUID will check the int to see if a value is provided via
// flags and convert it to uint32 and returns the values
func CheckUID(u int, g int) (uid, gid uint32) {
	if u == 0 {
		uid = uint32(0)
		gid = uint32(0)
	} else {
		uid = uint32(u)
		gid = uint32(g)
	}
	return uid, gid
}

// CheckMark will only print the check mark for progress bar
func CheckMark() func() {
	return func() {
		fmt.Printf("%s ✔ %s\n", "\033[32m", "\033[0m")
	}
}

// IsPathValid checks if a file or directory exists at the given path.
// It returns true if the path exists, false if it doesn't exist, and an error if any issue occurs during the check.
func IsPathValid(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, err
	} else {
		return false, err
	}
}

// TrimSAH will check the commit and if its not empty trim.
func TrimSHA(commitSHA string) string {
	if len(commitSHA) > 12 {
		shortSHA := commitSHA[:12]
		return shortSHA
	}
	return commitSHA
}
