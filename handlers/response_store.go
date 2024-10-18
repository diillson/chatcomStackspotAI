// handlers/response_store.go

package handlers

import (
	"github.com/chatcomStackspotAI/models"
	"sync"
)

type ResponseStore struct {
	mu        sync.RWMutex
	responses map[string]*models.ResponseData
}

func NewResponseStore() *ResponseStore {
	return &ResponseStore{
		responses: make(map[string]*models.ResponseData),
	}
}

func (store *ResponseStore) SetResponse(messageID string, data *models.ResponseData) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.responses[messageID] = data
}

func (store *ResponseStore) GetResponse(messageID string) (*models.ResponseData, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	data, exists := store.responses[messageID]
	return data, exists
}
