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

// Search returns karaoke results, ignoring queries that are too short.
func (s *CatalogService) Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return []domain.SearchResult{}, false, nil
	}
	if s.Catalog == nil {
		return nil, false, domain.ErrSearchFailed
	}
	results, stale, err := s.Catalog.Search(ctx, query)
	if err != nil {
		return nil, false, err
	}
	if results == nil {
		results = []domain.SearchResult{}
	}
	return results, stale, nil
}
