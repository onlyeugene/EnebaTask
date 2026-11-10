package products

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"eugene/internal/platform/httpx"
)

type Handler struct {
	svc    *Service
	search Searcher
}

func NewHandler(svc *Service, search Searcher) *Handler {
	return &Handler{svc: svc, search: search}
}

// list serves the store. With ?q= it goes through Atlas Search (typos, partial words);
// without it, a plain find on the B-tree indexes, which is cheaper and exact.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q, err := parseListQuery(r.URL.Query())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	var res ListResult
	if q.Q != "" {
		res, err = h.search.Search(r.Context(), q)
	} else {
		res, err = h.svc.List(r.Context(), q)
	}
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.OK(w, res)
}

func (h *Handler) filters(w http.ResponseWriter, r *http.Request) {
	f, err := h.svc.Filters(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.OK(w, f)
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	product, offers, err := h.svc.Detail(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		httpx.WriteError(w, r, notFound(err))
		return
	}
	httpx.OK(w, map[string]any{"product": product, "offers": offers})
}

func (h *Handler) cover(w http.ResponseWriter, r *http.Request) {
	svg, err := h.svc.CoverSVG(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		httpx.WriteError(w, r, notFound(err))
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	// An SVG opened directly in a tab can run scripts. Ours has none, but this header makes
	// sure the browser would refuse them anyway.
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(svg))
}

func notFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Product not found")
	}
	return err
}

// parseListQuery reads GET /api/products params (API.md). Any bad value is a 400 that names
// the param, so a typo in a link fails loudly instead of quietly ignoring the filter.
func parseListQuery(v url.Values) (ListQuery, error) {
	q := ListQuery{InStockOnly: true, Sort: SortBestselling, Page: 1, PageSize: 20}
	var err error

	if q.Q, err = SearchText(v.Get("q")); err != nil {
		return q, err
	}
	if q.Q != "" {
		q.Sort = SortRelevance // default with q; an explicit sort below still wins
	}

	if q.Types, err = enumList(v, "type", Types); err != nil {
		return q, err
	}
	if q.Platforms, err = enumList(v, "platform", Platforms); err != nil {
		return q, err
	}
	if q.Regions, err = enumList(v, "region", Regions); err != nil {
		return q, err
	}
	// Genres are free-form, so an unknown one just matches nothing.
	q.Genres = splitList(v.Get("genre"))

	if q.MinPriceCents, err = cents(v, "minPrice"); err != nil {
		return q, err
	}
	if q.MaxPriceCents, err = cents(v, "maxPrice"); err != nil {
		return q, err
	}
	if q.MinPriceCents != nil && q.MaxPriceCents != nil && *q.MinPriceCents > *q.MaxPriceCents {
		return q, badParam("minPrice", "must not be above maxPrice")
	}

	switch v.Get("inStock") {
	case "", "true":
	case "false":
		q.InStockOnly = false
	default:
		return q, badParam("inStock", "must be true or false")
	}

	if s := v.Get("sort"); s != "" {
		if !slices.Contains(Sorts, s) {
			return q, badParam("sort", "must be one of "+strings.Join(Sorts, ", "))
		}
		// relevance needs a search query; without one, keep the default (bestselling).
		if s != SortRelevance || q.Q != "" {
			q.Sort = s
		}
	}

	if q.Page, err = intParam(v, "page", 1, 1, 1_000_000); err != nil {
		return q, err
	}
	if q.PageSize, err = intParam(v, "pageSize", 20, 1, MaxPageSize); err != nil {
		return q, err
	}
	return q, nil
}

// SearchText trims the search box text and rejects anything over MaxQueryLen.
// Exported because the autocomplete handler applies the same rule.
func SearchText(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > MaxQueryLen {
		return "", badParam("q", fmt.Sprintf("must be at most %d characters", MaxQueryLen))
	}
	return s, nil
}

func enumList(v url.Values, name string, allowed []Option) ([]string, error) {
	values := splitList(v.Get(name))
	for _, val := range values {
		if !slices.ContainsFunc(allowed, func(o Option) bool { return o.Value == val }) {
			return nil, badParam(name, fmt.Sprintf("unknown value %q", val))
		}
	}
	return values, nil
}

// splitList turns "steam, PSN" into ["steam", "psn"]. Values are lowercased because all stored values are.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.ToLower(strings.TrimSpace(part)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func cents(v url.Values, name string) (*int64, error) {
	s := v.Get(name)
	if s == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return nil, badParam(name, "must be a whole number of cents, 0 or more")
	}
	return &n, nil
}

func intParam(v url.Values, name string, fallback, min, max int) (int, error) {
	s := v.Get(name)
	if s == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < min || n > max {
		return 0, badParam(name, fmt.Sprintf("must be a whole number from %d to %d", min, max))
	}
	return n, nil
}

func badParam(name, msg string) error {
	return httpx.NewError(http.StatusBadRequest, httpx.CodeBadRequest, name+": "+msg)
}
