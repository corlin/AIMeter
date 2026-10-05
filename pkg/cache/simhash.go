package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
	"math/bits"
	"strings"
	"unicode"
)

// ComputeExactHash returns the SHA-256 hex string of normalized text
func ComputeExactHash(text string) string {
	normalized := NormalizeText(text)
	hasher := sha256.New()
	hasher.Write([]byte(normalized))
	return hex.EncodeToString(hasher.Sum(nil))
}

// NormalizeText normalizes whitespace and lowers case
func NormalizeText(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))

	prevSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !prevSpace {
				builder.WriteRune(' ')
				prevSpace = true
			}
		} else {
			builder.WriteRune(unicode.ToLower(r))
			prevSpace = false
		}
	}

	return strings.TrimSpace(builder.String())
}

// Tokenize splits text into semantic tokens (CJK character unigrams/bigrams + Western words)
func Tokenize(text string) []string {
	normalized := NormalizeText(text)
	if normalized == "" {
		return nil
	}

	runes := []rune(normalized)
	tokens := make([]string, 0, len(runes))

	var currentWord strings.Builder

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Check if CJK character
		if unicode.Is(unicode.Han, r) {
			// Flush western word if any
			if currentWord.Len() > 0 {
				tokens = append(tokens, currentWord.String())
				currentWord.Reset()
			}

			// Add unigram
			tokens = append(tokens, string(r))

			// Add bigram if next rune exists and is also Han
			if i+1 < len(runes) && unicode.Is(unicode.Han, runes[i+1]) {
				tokens = append(tokens, string([]rune{r, runes[i+1]}))
			}
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else {
			// Punctuation or space
			if currentWord.Len() > 0 {
				tokens = append(tokens, currentWord.String())
				currentWord.Reset()
			}
		}
	}

	if currentWord.Len() > 0 {
		tokens = append(tokens, currentWord.String())
	}

	// Also add 3-grams for non-CJK subwords if available
	return tokens
}

// fnvHash64 returns a 64-bit hash of a string
func fnvHash64(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// ComputeSimHash computes a 64-bit SimHash fingerprint of the given text
func ComputeSimHash(text string) uint64 {
	tokens := Tokenize(text)
	if len(tokens) == 0 {
		return 0
	}

	var vector [64]int

	for _, token := range tokens {
		weight := 1
		runes := []rune(token)
		if len(runes) == 1 || !unicode.Is(unicode.Han, runes[0]) {
			weight = 2
		}

		h := fnvHash64(token)
		for i := 0; i < 64; i++ {
			bit := (h >> i) & 1
			if bit == 1 {
				vector[i] += weight
			} else {
				vector[i] -= weight
			}
		}
	}

	var simhash uint64
	for i := 0; i < 64; i++ {
		if vector[i] > 0 {
			simhash |= (1 << i)
		}
	}

	return simhash
}

// HammingDistance calculates the number of differing bits between two 64-bit hashes
func HammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// ComputeSimilarity computes the semantic similarity [0.0, 1.0] between two 64-bit SimHashes
func ComputeSimilarity(a, b uint64) float64 {
	dist := HammingDistance(a, b)
	sim := 1.0 - (float64(dist) / 64.0)
	if sim < 0.0 {
		return 0.0
	}
	return sim
}
