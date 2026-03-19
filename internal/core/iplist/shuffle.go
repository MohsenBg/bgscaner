package iplist

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
)

// Chunk size for in‑memory shuffling before writing to temporary files.
const chunkSize = 100_000

// chunkMeta holds metadata for each intermediate shuffled temp file.
type chunkMeta struct {
	path      string // temporary chunk file path
	remaining int    // number of lines in the chunk
}

// ShuffleFileFullyMemorySafe performs a full shuffle of an IP list file
// without loading the entire dataset into memory.
//
// Design:
//   - Reads active IPs via StreamActiveIPs()
//   - Splits them into fixed‑size chunks
//   - Each chunk is shuffled independently and stored on disk
//   - Finally, merges all chunks using weighted random selection.
//
// Returns: path to the final shuffled temporary file.
func ShuffleFileFullyMemorySafe(ctx context.Context, ipFile string) (string, error) {
	chunks := make([]chunkMeta, 0, 16) // pre‑allocate small slice; typical use has <16 chunks
	chunk := make([]string, 0, chunkSize)

	ips := make(chan string)
	go func() {
		defer close(ips)
		_ = StreamActiveIPs(ctx, ipFile, ips)
	}()

	// Process incoming IPs in chunks
	for ip := range ips {
		chunk = append(chunk, fmt.Sprintf("%s,1", ip))
		if len(chunk) >= chunkSize {
			meta, err := writeShuffledChunk(chunk)
			if err != nil {
				return "", fmt.Errorf("writing chunk: %w", err)
			}
			chunks = append(chunks, meta)
			chunk = make([]string, 0, chunkSize)
		}
	}

	// Flush remaining entries
	if len(chunk) > 0 {
		meta, err := writeShuffledChunk(chunk)
		if err != nil {
			return "", fmt.Errorf("writing final chunk: %w", err)
		}
		chunks = append(chunks, meta)
	}

	// Prepare final merged file
	finalFile, err := os.CreateTemp("", "shuffled_final_*.txt")
	if err != nil {
		return "", fmt.Errorf("create final temp: %w", err)
	}
	finalPath := finalFile.Name()
	finalFile.Close()

	if err := mergeChunksRandomlyWeighted(chunks, finalPath); err != nil {
		return "", fmt.Errorf("merging shuffled chunks: %w", err)
	}

	// Cleanup temporary chunk files
	for _, c := range chunks {
		_ = os.Remove(c.path)
	}

	return finalPath, nil
}

// writeShuffledChunk writes one chunk of shuffled IPs to a temporary file.
func writeShuffledChunk(chunk []string) (chunkMeta, error) {
	rand.Shuffle(len(chunk), func(i, j int) {
		chunk[i], chunk[j] = chunk[j], chunk[i]
	})

	tmpFile, err := os.CreateTemp("", "chunk_*.txt")
	if err != nil {
		return chunkMeta{}, fmt.Errorf("create temp chunk: %w", err)
	}
	defer tmpFile.Close()

	writer := bufio.NewWriterSize(tmpFile, 64*1024)
	for _, line := range chunk {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return chunkMeta{}, fmt.Errorf("write chunk: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return chunkMeta{}, fmt.Errorf("flush chunk writer: %w", err)
	}

	return chunkMeta{
		path:      tmpFile.Name(),
		remaining: len(chunk),
	}, nil
}

// mergeChunksRandomlyWeighted merges all shuffled chunk files into one final file
// using weighted random selection proportional to remaining items per chunk.
//
// Ensures uniform probability distribution across all source chunks
// while streaming line‑by‑line from disk (constant memory footprint).
func mergeChunksRandomlyWeighted(chunkMetas []chunkMeta, outputPath string) error {
	type chunkState struct {
		file      *os.File
		scanner   *bufio.Scanner
		remaining int
		current   string
	}

	var (
		chunks         []*chunkState
		totalRemaining int
	)

	// Initialize scanners for each chunk
	for _, meta := range chunkMetas {
		f, err := os.Open(meta.path)
		if err != nil {
			return fmt.Errorf("open chunk %s: %w", meta.path, err)
		}
		sc := bufio.NewScanner(f)
		if sc.Scan() {
			chunks = append(chunks, &chunkState{
				file:      f,
				scanner:   sc,
				remaining: meta.remaining,
				current:   sc.Text(),
			})
			totalRemaining += meta.remaining
		} else {
			f.Close() // empty chunk
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriterSize(outFile, 128*1024) // larger buffer for faster merge

	for totalRemaining > 0 {
		r := rand.Intn(totalRemaining)

		cumulative := 0
		var chosen *chunkState

		for _, ch := range chunks { // weighted random pick
			if ch.remaining <= 0 {
				continue
			}
			cumulative += ch.remaining
			if r < cumulative {
				chosen = ch
				break
			}
		}

		if chosen == nil {
			break // all exhausted
		}

		fmt.Fprintln(writer, chosen.current)

		totalRemaining--
		chosen.remaining--

		if chosen.scanner.Scan() {
			chosen.current = chosen.scanner.Text()
		} else {
			_ = chosen.file.Close()
			chosen.remaining = 0
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush output writer: %w", err)
	}

	return nil
}
