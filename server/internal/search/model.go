package search

// Suggestion is one row in the header search dropdown (API.md, /api/search/autocomplete).
type Suggestion struct {
	Slug               string `json:"slug"`
	Name               string `json:"name"`
	Type               string `json:"type"` // the row's pill badge
	Platform           string `json:"platform"`
	Region             string `json:"region"`
	OriginalPriceCents int64  `json:"originalPriceCents"` // struck through when discountPct > 0
	MinPriceCents      *int64 `json:"minPriceCents"`      // null = sold out
	DiscountPct        int    `json:"discountPct"`
	CoverURL           string `json:"coverUrl"`
}

// MaxSuggestions is how many rows the dropdown shows.
const MaxSuggestions = 8

// MinAutocompleteLen matches the index's minGrams: a single letter has no edgeGram to match.
const MinAutocompleteLen = 2
