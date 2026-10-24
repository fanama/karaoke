package usecase

import (
	"context"

	"fana-karaoke/internal/domain"
)

// QueueService owns songs, upvotes and queue ordering.
type QueueService struct {
	Deps
}

// NewQueueService builds the queue use case.
func NewQueueService(d Deps) *QueueService { return &QueueService{Deps: d} }

// AddSong resolves metadata, validates it, and appends the song.
func (s *QueueService) AddSong(ctx context.Context, code, deviceID, role, nickname, videoID, rawURL string) (*domain.Snapshot, error) {
	name, err := domain.CleanNickname(nickname)
	if err != nil {
		return nil, err
	}
	videoID, err = resolveVideoID(videoID, rawURL)
	if err != nil {
		return nil, err
	}
	if err := s.checkAdmission(code); err != nil {
		return nil, err
	}
	meta, err := s.lookup(ctx, videoID)
	if err != nil {
		return nil, err
	}
	return s.appendSong(code, deviceID, role, name, meta)
}

func resolveVideoID(videoID, rawURL string) (string, error) {
	if rawURL != "" {
		if domain.IsShortLink(rawURL) {
			return "", domain.ErrVideoShort
		}
		return domain.ParseYouTubeURL(rawURL)
	}
	if !domain.ValidVideoID(videoID) {
		return "", domain.ErrInvalidURL
	}
	return videoID, nil
}

func (s *QueueService) checkAdmission(code string) error {
	room, ok := s.Rooms.Get(code)
	if !ok {
		return domain.ErrRoomNotFound
	}
	if room.QueueLocked {
		return domain.ErrQueueLocked
	}
	if len(room.Entries) >= s.Cfg.MaxQueue {
		return domain.ErrQueueFull
	}
	return nil
}

func (s *QueueService) lookup(ctx context.Context, videoID string) (domain.VideoMeta, error) {
	if s.Catalog == nil {
		return domain.VideoMeta{}, domain.ErrSearchFailed
	}
	meta, err := s.Catalog.Lookup(ctx, videoID)
	if err != nil {
		return domain.VideoMeta{}, err
	}
	if err := domain.ValidateVideoMeta(meta); err != nil {
		return domain.VideoMeta{}, err
	}
	if meta.ThumbnailURL == "" {
		meta.ThumbnailURL = domain.ThumbnailURL(meta.VideoID)
	}
	if err := s.rememberSong(ctx, meta); err != nil {
		return domain.VideoMeta{}, err
	}
	return meta, nil
}

// rememberSong keeps every played or pasted song in the local library.
func (s *QueueService) rememberSong(ctx context.Context, meta domain.VideoMeta) error {
	if s.Library == nil {
		return nil
	}
	song := domain.SearchResult{
		VideoID:         meta.VideoID,
		Title:           meta.Title,
		Channel:         meta.Channel,
		ThumbnailURL:    meta.ThumbnailURL,
		DurationSeconds: meta.DurationSeconds,
	}
	return s.Library.Save(ctx, []domain.SearchResult{song})
}

func (s *QueueService) appendSong(code, deviceID, role, nickname string, meta domain.VideoMeta) (*domain.Snapshot, error) {
	entry := domain.QueueEntry{
		ID:              s.IDs.NewID("q_"),
		VideoID:         meta.VideoID,
		Title:           meta.Title,
		Channel:         meta.Channel,
		ThumbnailURL:    meta.ThumbnailURL,
		DurationSeconds: meta.DurationSeconds,
		Nickname:        nickname,
		DeviceID:        deviceID,
		Votes:           map[string]bool{},
	}
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.EnqueueSong(room, entry, s.Now(), s.Cfg.MaxQueue)
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}

// ToggleVote flips the viewer's upvote on a queued song.
func (s *QueueService) ToggleVote(code, deviceID, role, entryID string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.ToggleEntryVote(room, entryID, deviceID)
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}

// RemoveSong drops a queued song (host only at the HTTP layer).
func (s *QueueService) RemoveSong(code, entryID, deviceID, role string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.RemoveEntry(room, entryID, s.Now())
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}

// Reorder applies a host-provided queue order.
func (s *QueueService) Reorder(code string, ids []string, deviceID, role string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.ReorderQueue(room, ids, s.Now())
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}

// ClearQueue empties the queue.
func (s *QueueService) ClearQueue(code, deviceID, role string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.ClearEntries(room, s.Now()), nil
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}

// LockQueue prevents or allows new songs.
func (s *QueueService) LockQueue(code string, locked bool, deviceID, role string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.SetQueueLock(room, locked, s.Now()), nil
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}
