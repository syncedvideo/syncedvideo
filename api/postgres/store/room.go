package store

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/syncedvideo/syncedvideo"
)

// RoomStore implements syncedvideo.RoomStore
type RoomStore struct {
	db *sqlx.DB
}

func (s *RoomStore) Get(id uuid.UUID) (syncedvideo.Room, error) {
	r := syncedvideo.Room{}
	err := s.db.Get(&r, "SELECT * FROM sv_room AS room WHERE room.id = $1 LIMIT 1", id)
	if err != nil {
		return syncedvideo.Room{}, fmt.Errorf("error getting room: %w", err)
	}
	items, err := s.GetPlaylistItems(id)
	if err != nil {
		return syncedvideo.Room{}, fmt.Errorf("error getting room playlist items: %s", err)
	}
	r.PlaylistItems = items
	return r, nil
}

func (s *RoomStore) Create(r *syncedvideo.Room) error {
	var ownerID *uuid.UUID
	if r.OwnerUserID != uuid.Nil {
		ownerID = &r.OwnerUserID
	}
	err := s.db.Get(r, "INSERT INTO sv_room VALUES ($1, $2, $3, $4) RETURNING *", r.ID, ownerID, r.Name, r.Description)
	if err != nil {
		return fmt.Errorf("error creating room: %w", err)
	}
	return nil
}

func (s *RoomStore) Update(r *syncedvideo.Room) error {
	err := s.db.Get(r, "UPDATE sv_room SET name=$1, owner_user_id=$2 WHERE id=$3 RETURNING *", r.Name, r.OwnerUserID, r.ID)
	if err != nil {
		return fmt.Errorf("error updating room: %w", err)
	}
	return nil
}

func (s *RoomStore) Delete(id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM sv_room WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("error deleting room: %w", err)
	}
	return nil
}

func (s *RoomStore) GetPlaylistItem(r *syncedvideo.Room, id uuid.UUID) (syncedvideo.PlaylistItem, error) {
	item := syncedvideo.PlaylistItem{}
	err := s.db.Get(&item, "SELECT * FROM sv_playlist_item WHERE room_id=$1 AND item_id=$2", r.ID, id)
	if err != nil {
		return syncedvideo.PlaylistItem{}, fmt.Errorf("error getting playlist item: %w", err)
	}
	return item, nil
}

func (s *RoomStore) GetPlaylistItems(roomID uuid.UUID) (map[uuid.UUID]syncedvideo.PlaylistItem, error) {
	rows, err := s.db.Query(`
		SELECT item.id, item.room_id, item.user_id, vote.id, vote.item_id, vote.user_id 
		FROM sv_playlist_item item
		LEFT JOIN sv_playlist_item_vote vote
		ON item.id = vote.item_id
		WHERE item.room_id = $1
	`, roomID)
	if err != nil {
		return nil, fmt.Errorf("error getting all playlist items: %w", err)
	}

	items := make(map[uuid.UUID]syncedvideo.PlaylistItem)
	for rows.Next() {
		item := syncedvideo.PlaylistItem{}
		vote := syncedvideo.PlaylistItemVote{}
		err := rows.Scan(&item.ID, &item.RoomID, &item.UserID, &vote.ID, &vote.ItemID, &vote.UserID)
		if err != nil {
			return nil, fmt.Errorf("error scanning room playlist item: %w", err)
		}
		if vote.ID != uuid.Nil {
			item.Votes = append(item.Votes, vote)
		}
		if _, exists := items[item.ID]; !exists {
			items[item.ID] = item
		}
	}

	return items, nil
}

func (s *RoomStore) CreatePlaylistItem(r *syncedvideo.Room, p *syncedvideo.PlaylistItem) error {
	err := s.db.Get(p, "INSERT INTO sv_room_playlist_item VALUES ($1, $2, $3) RETURNING *", p.ID, r.ID, p.UserID)
	if err != nil {
		return fmt.Errorf("error creating playlist item: %w", err)
	}
	return nil
}

func (s *RoomStore) UpdatePlaylistItem(r *syncedvideo.Room, p *syncedvideo.PlaylistItem) error {
	err := s.db.Get(p, "UPDATE sv_room_playlist_item SET room_id=$1, user_id=$2 WHERE id=$3 RETURNING *", r.ID, p.UserID, p.ID)
	if err != nil {
		return fmt.Errorf("error updating playlist item: %w", err)
	}
	return nil
}

func (s *RoomStore) DeletePlaylistItem(r *syncedvideo.Room, id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM sv_room_playlist_item WHERE id=$1 AND room_id=$2", id, r.ID)
	if err != nil {
		return fmt.Errorf("error deleting playlist item: %w", err)
	}
	return nil
}