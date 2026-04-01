package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const ffmpegPath = "ffmpeg"

type ansi struct {
	reset, bold, dim                        string
	cyan, green, yellow, red, blue, magenta string
}

func main() {
	a := NewANSI()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("%s%sFLAC → AAC%s  %s(320 kbps, M4A · Apple Music friendly)%s\n\n",
		a.bold+a.magenta, "", a.reset, a.dim, a.reset)

	fmt.Printf("%s① Input directory%s  %s(where your .flac files live)%s\n",
		a.bold+a.cyan, a.reset, a.dim, a.reset)
	fmt.Printf("   %s›%s ", a.green, a.reset)
	scanner.Scan()
	inputRoot := NormalizeUserPath(scanner.Text())

	fmt.Printf("\n%s② Output directory%s  %s(mirrored folder structure · .m4a files here)%s\n",
		a.bold+a.cyan, a.reset, a.dim, a.reset)
	fmt.Printf("   %s›%s ", a.green, a.reset)
	scanner.Scan()
	outputRoot := NormalizeUserPath(scanner.Text())

	inputRoot, err := filepath.Abs(inputRoot)
	if err != nil {
		fatal(a, "resolve input path: %v", err)
	}

	outputRoot, err = filepath.Abs(outputRoot)
	if err != nil {
		fatal(a, "resolve output path: %v", err)
	}

	info, err := os.Stat(inputRoot)
	if err != nil {
		fatal(a, "stat input directory: %v", err)
	}

	if !info.IsDir() {
		fatal(a, "input path is not a directory: %s", inputRoot)
	}

	err = os.MkdirAll(outputRoot, 0o755)
	if err != nil {
		fatal(a, "create output directory: %v", err)
	}

	var count int

	err = filepath.WalkDir(inputRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || !IsFlacFile(path) {
			return nil
		}

		outPath, err := OutputPathForFLAC(inputRoot, outputRoot, path)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(outPath), 0o755)
		if err != nil {
			return err
		}

		fmt.Printf("%s● Converting%s\n", a.yellow+a.bold, a.reset)
		fmt.Printf("  %s%s%s\n", a.dim, path, a.reset)
		fmt.Printf("  %s%s→%s %s%s\n", a.dim, a.blue, a.reset, a.green, outPath)

		cmd := exec.Command(ffmpegPath,
			"-hide_banner", "-loglevel", "error", "-nostats",
			"-y", "-i", path,
			"-vn",
			"-c:a", "aac", "-b:a", "320k",
			"-movflags", "+faststart",
			outPath,
		)

		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("ffmpeg %s: %w", path, err)
		}

		count++

		fmt.Println()

		return nil
	})

	if err != nil {
		fatal(a, "%v", err)
	}

	fmt.Printf("%s✓ Done.%s  Converted %s%d%s file(s).\n",
		a.bold+a.green, a.reset, a.bold, count, a.reset)
}

func NormalizeUserPath(s string) string {
	s = strings.TrimSpace(s)
	if runtime.GOOS == "windows" {
		if strings.Contains(s, `\ `) || strings.Contains(s, `\(`) || strings.Contains(s, `\[`) {
			return UnescapeShellPath(s)
		}

		return s
	}

	return UnescapeShellPath(s)
}

func UnescapeShellPath(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) {
			if s[i+1] == '\n' {
				i += 2

				continue
			}

			b.WriteByte(s[i+1])
			i += 2

			continue
		}

		b.WriteByte(s[i])
		i++
	}

	return b.String()
}

func IsFlacFile(path string) bool {
	ext := filepath.Ext(path)
	return strings.EqualFold(ext, ".flac")
}

func OutputPathForFLAC(inputRoot, outputRoot, flacPath string) (string, error) {
	inputRoot = filepath.Clean(inputRoot)
	flacPath = filepath.Clean(flacPath)
	rel, err := filepath.Rel(inputRoot, flacPath)
	if err != nil {
		return "", err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("flac path %q is not under input root %q", flacPath, inputRoot)
	}

	base := strings.TrimSuffix(rel, filepath.Ext(rel))

	return filepath.Join(outputRoot, base+".m4a"), nil
}

func ColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

func NewANSI() ansi {
	if !ColorEnabled() {
		return ansi{}
	}

	return ansi{
		reset: "\033[0m", bold: "\033[1m", dim: "\033[2m",
		cyan: "\033[36m", green: "\033[32m", yellow: "\033[33m",
		red: "\033[31m", blue: "\033[34m", magenta: "\033[35m",
	}
}

func fatal(a ansi, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s%s%s\n", a.red+a.bold, msg, a.reset)
	os.Exit(1)
}
