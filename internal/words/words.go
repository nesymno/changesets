package words

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

var adjectives = []string{
	"angry", "brave", "calm", "dark", "eager",
	"fair", "glad", "happy", "icy", "jolly",
	"keen", "lazy", "mild", "neat", "odd",
	"pink", "quick", "rare", "safe", "tall",
	"ugly", "vast", "warm", "young", "zany",
	"bold", "cool", "deep", "easy", "fast",
	"gold", "hard", "iron", "just", "kind",
	"lean", "mean", "nice", "open", "pale",
	"rich", "slim", "thin", "used", "wide",
	"able", "bare", "cute", "dull", "even",
	"fine", "gray", "high", "idle", "jade",
	"lame", "lost", "mass", "new", "old",
	"poor", "raw", "shy", "tiny", "wet",
	"apt", "big", "dry", "fit", "grim",
	"hot", "ill", "low", "mad", "red",
	"sad", "tan", "dim", "flat", "full",
	"green", "huge", "long", "loud", "pure",
	"real", "ripe", "soft", "sour", "tame",
	"true", "vain", "wild", "wise", "worn",
	"aged", "blue", "crisp", "dense", "fresh",
	"fuzzy", "sharp", "stark", "steep", "swift",
}

var nouns = []string{
	"ant", "bear", "cat", "deer", "elk",
	"fox", "goat", "hawk", "ibis", "jay",
	"kite", "lynx", "moth", "newt", "owl",
	"puma", "quail", "robin", "seal", "toad",
	"vole", "wolf", "yak", "bass", "crab",
	"dove", "frog", "hare", "lark", "mole",
	"pike", "wren", "orca", "swan", "wasp",
	"crow", "dusk", "dawn", "echo", "fire",
	"gale", "haze", "jade", "lake", "mist",
	"noon", "opal", "peak", "rain", "snow",
	"tide", "vale", "wave", "zinc", "arch",
	"bolt", "clay", "dome", "edge", "fern",
	"glow", "hill", "iron", "knot", "leaf",
	"moss", "nest", "pond", "reef", "sand",
	"tree", "vine", "wood", "bark", "beam",
	"cape", "dale", "elm", "flint", "glen",
	"herb", "isle", "kern", "lime", "mint",
	"oak", "pine", "rose", "sage", "thorn",
	"ash", "bay", "dew", "fig", "gem",
	"hops", "ink", "jet", "key", "log",
}

// Generate creates a random slug in the format "adj-adj-noun".
// It checks for collisions with existing files in changesDir.
func Generate(changesDir string) (string, error) {
	for attempts := 0; attempts < 100; attempts++ {
		adj1, err := randomElement(adjectives)
		if err != nil {
			return "", err
		}
		adj2, err := randomElement(adjectives)
		if err != nil {
			return "", err
		}
		noun, err := randomElement(nouns)
		if err != nil {
			return "", err
		}

		slug := fmt.Sprintf("%s-%s-%s", adj1, adj2, noun)
		filename := slug + ".md"

		path := filepath.Join(changesDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return slug, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique slug after 100 attempts")
}

func randomElement(slice []string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	return slice[n.Int64()], nil
}

// SlugToFilename converts a slug to a markdown filename.
func SlugToFilename(slug string) string {
	return slug + ".md"
}

// FilenameToSlug converts a markdown filename back to a slug.
func FilenameToSlug(filename string) string {
	return strings.TrimSuffix(filename, ".md")
}
