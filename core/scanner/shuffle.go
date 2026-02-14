package scanner

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

// chunk size for memory-efficient shuffle
const chunkSize = 100_000

// ShuffleFileFullyMemorySafe fully shuffles a large file in a memory-efficient way.
// Works with millions of lines without loading everything into memory.
func ShuffleFileFullyMemorySafe(inputPath string) (string, error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}
	defer inputFile.Close()

	// Step 1: split into shuffled chunks
	var chunkFiles []string
	scanner := bufio.NewScanner(inputFile)
	chunk := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		chunk = append(chunk, line)

		if len(chunk) >= chunkSize {
			tmpFile, err := writeShuffledChunk(chunk)
			if err != nil {
				return "", err
			}
			chunkFiles = append(chunkFiles, tmpFile)
			chunk = nil
		}
	}

	// flush remaining lines
	if len(chunk) > 0 {
		tmpFile, err := writeShuffledChunk(chunk)
		if err != nil {
			return "", err
		}
		chunkFiles = append(chunkFiles, tmpFile)
	}

	// Step 2: merge chunks randomly into final file
	finalFile, err := os.CreateTemp("", "shuffled_final_*.txt")
	if err != nil {
		return "", err
	}
	defer finalFile.Close()

	if err := mergeChunksRandomly(chunkFiles, finalFile.Name()); err != nil {
		return "", err
	}

	// delete temp chunk files
	for _, f := range chunkFiles {
		os.Remove(f)
	}

	return finalFile.Name(), nil
}

// writeShuffledChunk shuffles a chunk in memory and writes to a temp file
func writeShuffledChunk(chunk []string) (string, error) {
	rand.Shuffle(len(chunk), func(i, j int) { chunk[i], chunk[j] = chunk[j], chunk[i] })

	tmpFile, err := os.CreateTemp("", "chunk_*.txt")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	writer := bufio.NewWriter(tmpFile)
	for _, line := range chunk {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()

	return tmpFile.Name(), nil
}

// mergeChunksRandomly merges multiple chunk files into one fully shuffled file
func mergeChunksRandomly(chunkFiles []string, outputPath string) error {
	type fileScanner struct {
		file    *os.File
		scanner *bufio.Scanner
	}

	var scanners []fileScanner
	for _, path := range chunkFiles {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		scanners = append(scanners, fileScanner{
			file:    f,
			scanner: bufio.NewScanner(f),
		})
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	// load first line from each scanner
	lines := make([]string, len(scanners))
	valid := make([]bool, len(scanners))
	for i, s := range scanners {
		if s.scanner.Scan() {
			lines[i] = s.scanner.Text()
			valid[i] = true
		}
	}

	// repeatedly pick a random available line
	for {
		available := 0
		for _, v := range valid {
			if v {
				available++
			}
		}
		if available == 0 {
			break // all files exhausted
		}

		// pick random available scanner
		var idx int
		for {
			idx = rand.Intn(len(scanners))
			if valid[idx] {
				break
			}
		}

		fmt.Fprintln(writer, lines[idx])

		// advance scanner
		if scanners[idx].scanner.Scan() {
			lines[idx] = scanners[idx].scanner.Text()
		} else {
			valid[idx] = false
			scanners[idx].file.Close()
		}
	}

	writer.Flush()
	return nil
}
