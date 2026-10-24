package usecase

import (
	"context"
	"strings"

	"fana-karaoke/internal/domain"
)

// Facade is the single driving port for HTTP and WebSocket adapters.
type Facade struct {
	*RoomService
	*QueueService
	*PlaybackService
	*VoteService
	*CatalogService
	Lifecycle *LifecycleService
}

// NewFacade wires every use case over the same dependencies.
func NewFacade(d Deps) *Facade {
	return &Facade{
		RoomService:     NewRoomService(d),
		QueueService:    NewQueueService(d),
		PlaybackService: NewPlaybackService(d),
		VoteService:     NewVoteService(d),
		CatalogService:  NewCatalogService(d),
		Lifecycle:       NewLifecycleService(d),
	}
}

// CatalogService answers search queries.
type CatalogService struct {
	Deps
}

// NewCatalogService builds the catalog use case.
func NewCatalogService(d Deps) *CatalogService { return &CatalogService{Deps: d} }

// Search answers from the local library first, then falls back online. Online
// hits are remembered so later searches do not spend the upstream quota.
func (s *CatalogService) Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return []domain.SearchResult{}, false, nil
	}
	if s.Catalog == nil {
		return nil, false, domain.ErrSearchFailed
	}
	if local := s.localSearch(ctx, query); len(local) > 0 {
		return local, false, nil
	}
	results, stale, err := s.Catalog.Search(ctx, query)
	if err != nil {
		return nil, false, err
	}
	if results == nil {
		results = []domain.SearchResult{}
	}
	s.remember(ctx, results)
	return results, stale, nil
}

func (s *CatalogService) localSearch(ctx context.Context, query string) []domain.SearchResult {
	if s.Library == nil {
		return nil
	}
	results, err := s.Library.Search(ctx, query, 20)
	if err != nil {
		return nil
	}
	return results
}

func (s *CatalogService) remember(ctx context.Context, results []domain.SearchResult) {
	if s.Library == nil || len(results) == 0 {
		return
	}
	_ = s.Library.Save(ctx, results)
}
