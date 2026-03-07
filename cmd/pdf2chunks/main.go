// pdf2chunks extracts text from a PDF via pdftotext (poppler) and outputs
// SQL for lesson_content_chunks. No Python or Go PDF libs required.
//
// Install: brew install poppler  (macOS)
// Usage: pdftotext -layout - book.pdf | go run ./cmd/pdf2chunks --lesson-id N
// Or: go run ./cmd/pdf2chunks --lesson-id N --pdf path/to/book.pdf
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"
)

const defaultChunkSize = 3500

func main() {
	lessonID := flag.Int("lesson-id", 0, "lesson_id for generated INSERTs (required for SQL output)")
	chunkSize := flag.Int("chunk-size", defaultChunkSize, "max runes per chunk")
	pdfPath := flag.String("pdf", "", "path to PDF (runs pdftotext internally)")
	flag.Parse()

	if *lessonID <= 0 {
		fmt.Fprintf(os.Stderr, "usage: pdftotext -layout - your.pdf | %s --lesson-id N [--chunk-size 3500]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  or: %s --lesson-id N --pdf path/to.pdf\n", os.Args[0])
		os.Exit(1)
	}

	var text string
	if *pdfPath != "" {
		cmd := exec.Command("pdftotext", "-layout", *pdfPath, "-")
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pdftotext failed (install: brew install poppler): %v\n", err)
			os.Exit(1)
		}
		text = string(out)
	} else if path := flag.Arg(0); path != "" {
		cmd := exec.Command("pdftotext", "-layout", path, "-")
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pdftotext failed: %v\n", err)
			os.Exit(1)
		}
		text = string(out)
	} else {
		// Read from stdin (piped from pdftotext)
		sc := bufio.NewScanner(os.Stdin)
		var b strings.Builder
		for sc.Scan() {
			b.WriteString(sc.Text())
			b.WriteByte('\n')
		}
		if err := sc.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
			os.Exit(1)
		}
		text = b.String()
	}

	chunks := splitIntoChunks(normalizeText(text), *chunkSize)
	fmt.Printf("-- %d chunks from PDF for lesson_id %d\n", len(chunks), *lessonID)
	fmt.Println("INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text) VALUES")
	for i, c := range chunks {
		if i > 0 {
			fmt.Print(",\n")
		}
		fmt.Printf("  (%d, %d, %s)", *lessonID, i+1, quoteSQL(c))
	}
	fmt.Println("\nON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;")
}

func normalizeText(s string) string {
	// Remove page markers like "-- 10 of 310 --" and "SRE: Коллективный разум" headers
	rePage := regexp.MustCompile(`(?m)^--\s*\d+\s+of\s+\d+\s+--\s*$`)
	reHeader := regexp.MustCompile(`(?m)^SRE:\s*Коллективный разум\s*$`)
	s = rePage.ReplaceAllString(s, "")
	s = reHeader.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}

func splitIntoChunks(text string, maxRunes int) []string {
	var chunks []string
	paragraphs := splitParagraphs(text)
	var current strings.Builder
	currentLen := 0
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		pr := utf8.RuneCountInString(p) + 1 // +1 for newline
		if currentLen > 0 && currentLen+pr > maxRunes {
			chunks = append(chunks, strings.TrimRight(current.String(), "\n"))
			current.Reset()
			currentLen = 0
		}
		current.WriteString(p)
		current.WriteByte('\n')
		currentLen += pr
	}
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimRight(current.String(), "\n"))
	}
	return chunks
}

func splitParagraphs(text string) []string {
	return regexp.MustCompile(`\n\s*\n`).Split(text, -1)
}

func quoteSQL(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
