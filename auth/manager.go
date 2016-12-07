package auth

import (
	"bytes"
	"errors"
	"log"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	cache "github.com/patrickmn/go-cache"

	storage "google.golang.org/api/storage/v1"
)

// Manager provides cached authentication from GCS via user name and token
type Manager struct {
	authCache *cache.Cache
	gcs       *storage.Service
	gcsBucket string
}

// NewManager connects to GCS
func NewManager(ctx context.Context, gcsBucket string) (*Manager, error) {
	if gcsBucket == "" {
		log.Println("Authentication is disabled, no GCS bucket with authentication data provided")
		return &Manager{}, nil
	}
	am := Manager{
		// 5 minute TTL, purge every 30 seconds.
		authCache: cache.New(5*time.Minute, 30*time.Second),
		gcsBucket: gcsBucket,
	}
	client, err := google.DefaultClient(ctx, storage.DevstorageReadOnlyScope)
	if err != nil {
		return nil, err
	}
	am.gcs, err = storage.New(client)
	return &am, err
}

func (am *Manager) fetchAuthToken(user string) (string, error) {
	r, err := am.gcs.Objects.Get(am.gcsBucket, user).Download()
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	_, err = b.ReadFrom(r.Body)
	// TODO(dotdoom): encode files in JSON
	return string(b.Bytes()), err
}

func (am *Manager) getAuthToken(user string) (string, error) {
	token, found := am.authCache.Get(user)
	if found {
		return token.(string), nil
	}
	tokenString, err := am.fetchAuthToken(user)
	if err == nil {
		am.authCache.Set(user, tokenString, cache.DefaultExpiration)
	}
	return tokenString, err
}

// CheckAccess returns nil if access is granted
func (am *Manager) CheckAccess(user, token string) error {
	// TODO(dotdoom): add 3rd parameter, level

	var (
		actualToken string
		err         error
	)
	if am.gcs == nil {
		actualToken, err = am.getAuthToken(user)
		if err != nil {
			return err
		}
		if actualToken == "" {
			return errors.New("Cannot verify the token")
		}
	} else {
		actualToken = ""
	}

	if actualToken != token {
		return errors.New("Incorrect token supplied")
	}
	return nil
}
