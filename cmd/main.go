package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"contractlens/internal/openapi"
)

const compareUsage = "usage: contractlens compare -base old.yaml -head new.yaml"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		// Use a nonzero status so scripts can detect invalid or unreadable inputs.
		os.Exit(1)
	}
}

// run validates the two input specifications and prints only version and path counts.
func run(args []string, output io.Writer) error {
	if len(args) == 0 || args[0] != "compare" {
		return errors.New(compareUsage)
	}

	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	basePath := flags.String("base", "", "path to the original OpenAPI specification")
	headPath := flags.String("head", "", "path to the updated OpenAPI specification")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, writeErr := fmt.Fprintln(output, compareUsage)
			return writeErr
		}
		return errors.New("invalid compare options; " + compareUsage)
	}

	if flags.NArg() != 0 {
		return errors.New("unexpected argument; " + compareUsage)
	}

	if *basePath == "" || *headPath == "" {
		return errors.New("both -base and -head are required; " + compareUsage)
	}

	baseData, err := readSpec(*basePath)
	if err != nil {
		return fmt.Errorf("could not read base specification: %w", err)
	}
	headData, err := readSpec(*headPath)
	if err != nil {
		return fmt.Errorf("could not read updated specification: %w", err)
	}

	base, err := openapi.Parse(baseData)
	if err != nil {
		return fmt.Errorf("base specification is invalid: %w", err)
	}
	head, err := openapi.Parse(headData)
	if err != nil {
		return fmt.Errorf("updated specification is invalid: %w", err)
	}

	if _, err := fmt.Fprintf(output, "base: OpenAPI %q (%d paths)\n", base.Version, len(base.Paths)); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	if _, err := fmt.Fprintf(output, "head: OpenAPI %q (%d paths)\n", head.Version, len(head.Paths)); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}

	if _, err := fmt.Fprintln(output, "Both specifications are valid; change detection is not implemented yet."); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	return nil
}

// readSpec reads one file while enforcing the parser's size limit during I/O.
func readSpec(path string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, errors.New("input directory could not be opened")
	}

	defer root.Close()

	file, err := root.Open(filepath.Base(path))
	if err != nil {
		return nil, errors.New("file could not be opened")
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("input must be a regular file")
	}
	if info.Size() > int64(openapi.MaxDocumentBytes) {
		return nil, fmt.Errorf("file exceeds the %d-byte limit", openapi.MaxDocumentBytes)
	}

	limited := io.LimitReader(file, int64(openapi.MaxDocumentBytes)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, errors.New("file could not be read")
	}

	// Again checking because what if in between of file.stat and readAll file size increase so we check it again in the end.
	if len(data) > openapi.MaxDocumentBytes {
		return nil, fmt.Errorf("file exceeds the %d-byte limit", openapi.MaxDocumentBytes)
	}

	return data, nil
}
