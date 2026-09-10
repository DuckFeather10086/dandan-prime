//go:build !js && !wasm
// +build !js,!wasm

package matchusecase

import (
	"bufio"
	"os"
	"testing"
)

// TestEpisodeNumberCorpus runs the parser over a newline-separated list of
// real file names and reports what it could not read. It is skipped unless
// DANDAN_PARSE_CORPUS points at such a file:
//
//	sqlite3 cmd/media_library.db 'select file_name from episode_infos' > /tmp/corpus.txt
//	DANDAN_PARSE_CORPUS=/tmp/corpus.txt go test ./usecase/matchUseCase/ -run Corpus -v
//
// It asserts nothing about the rate -- extras and single-film directories
// legitimately have no episode number. It is here to make the misses
// inspectable when the parser changes.
func TestEpisodeNumberCorpus(t *testing.T) {
	path := os.Getenv("DANDAN_PARSE_CORPUS")
	if path == "" {
		t.Skip("set DANDAN_PARSE_CORPUS to a file of newline-separated file names")
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var total, matched, extras int
	var misses []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		name := scanner.Text()
		if name == "" {
			continue
		}
		total++
		if IsExtra(name) {
			extras++
			continue
		}
		if _, ok := EpisodeNumber(name); ok {
			matched++
		} else if len(misses) < 60 {
			misses = append(misses, name)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	features := total - extras
	t.Logf("total=%d extras=%d features=%d numbered=%d (%.1f%%)",
		total, extras, features, matched, 100*float64(matched)/float64(max(features, 1)))
	for _, m := range misses {
		t.Logf("  no episode number: %s", m)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
