package search

// Suggestion is one row in the header search dropdown (API.md, /api/search/autocomplete).
type Suggestion struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Platform      string `json:"platform"`
	Region        string `json:"region"`
	MinPriceCents *int64 `json:"minPriceCents"` // null = sold out
	CoverURL      string `json:"coverUrl"`
}

// MaxSuggestions is how many rows the dropdown shows.
const MaxSuggestions = 8

// MinAutocompleteLen matches the index's minGrams: a single letter has no edgeGram to match.
const MinAutocompleteLen = 2
