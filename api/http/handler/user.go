package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/syncedvideo/syncedvideo"
	"github.com/syncedvideo/syncedvideo/http/response"
)

type userHandler struct {
	store syncedvideo.Store
}

func RegisterUserHandler(router chi.Router, store syncedvideo.Store) {
	userHandler := newUserHandler(store)
	router.Route("/user", func(router2 chi.Router) {
		router2.Post("/auth", userHandler.Auth)
	})
}

func newUserHandler(s syncedvideo.Store) *userHandler {
	return &userHandler{
		store: s,
	}
}

const userCookieKey string = "userID"

func hasUserCookie(r *http.Request) bool {
	c, err := r.Cookie(userCookieKey)
	if err != nil {
		return false
	}
	return c.Value != ""
}

func getUserFromCookie(r *http.Request, userStore syncedvideo.UserStore) (syncedvideo.User, error) {
	userIDCookie, err := r.Cookie(userCookieKey)
	if err != nil {
		return syncedvideo.User{}, err
	}
	if userIDCookie.Value == "" {
		return syncedvideo.User{}, errors.New(userCookieKey + " cookie value is empty")
	}

	userID, err := uuid.Parse(userIDCookie.Value)
	if err != nil {
		return syncedvideo.User{}, err
	}
	if userID == uuid.Nil {
		return syncedvideo.User{}, errors.New("userID is nil")
	}

	user, err := userStore.Get(userID)
	if err != nil {
		return syncedvideo.User{}, err
	}
	return user, nil
}

func (h *userHandler) Auth(w http.ResponseWriter, r *http.Request) {
	user := syncedvideo.User{}

	if hasUserCookie(r) {
		u, err := getUserFromCookie(r, h.store.User())
		if err != nil && err != sql.ErrNoRows {
			log.Printf("error getting user: %v", err)
			response.WithError(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		user = u
	}

	if user.ID == uuid.Nil {
		err := h.store.User().Create(&user)
		if err != nil {
			log.Printf("error creating user: %v", err)
			response.WithError(w, "something went wrong", http.StatusInternalServerError)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     userCookieKey,
		Value:    user.ID.String(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Expires:  time.Now().UTC().Add(24 * time.Hour * 30), // 30 days
	})

	response.WithJSON(w, user, http.StatusOK)
}