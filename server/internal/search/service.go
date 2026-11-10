package search

import (
	"context"

	"eugene/internal/products"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Search implements products.Searcher, so /api/products?q= returns the same shape as the plain list.
func (s *Service) Search(ctx context.Context, q products.ListQuery) (products.ListResult, error) {
	found, total, err := s.repo.Search(ctx, q)
	if err != nil {
		return products.ListResult{}, err
	}
	items := make([]products.ProductCard, 0, len(found)) // [] not null in JSON when nothing matches
	for _, p := range found {
		items = append(items, products.ToCard(p))
	}
	return products.ListResult{Items: items, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

// Autocomplete returns up to MaxSuggestions for the header search box.
// Text shorter than MinAutocompleteLen returns none without asking the database.
func (s *Service) Autocomplete(ctx context.Context, text string) ([]Suggestion, error) {
	out := []Suggestion{}
	if len([]rune(text)) < MinAutocompleteLen {
		return out, nil
	}
	found, err := s.repo.Autocomplete(ctx, text, MaxSuggestions)
	if err != nil {
		return nil, err
	}
	for _, p := range found {
		card := products.ToCard(p) // reuse ToCard so coverUrl is built in one place
		out = append(out, Suggestion{
			Slug: card.Slug, Name: card.Name, Platform: card.Platform, Region: card.Region,
			MinPriceCents: card.MinPriceCents, CoverURL: card.CoverURL,
		})
	}
	return out, nil
}
