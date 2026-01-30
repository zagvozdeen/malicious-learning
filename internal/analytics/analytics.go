package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"maps"
	"strconv"
	"sync"
	"sync/atomic"
)

type Metrics interface {
	Snapshot() Snapshot

	AppUsersCreatedCountInc()
	AppNotMessageUpdateCountInc()
	AppGeneratedRecommendationsCountInc()
	AppUpdatedUserAnswersCountInc()
	AppCreatedTestSessionsCountInc()
	AppResponsesTotalInc(path string, code int)
}

type Snapshot struct {
	AppUsersCreatedCount             int64            `json:"app_users_created_count"`
	AppNotMessageUpdateCount         int64            `json:"app_not_message_update_count"`
	AppGeneratedRecommendationsCount int64            `json:"app_generated_recommendations_count"`
	AppUpdatedUserAnswersCount       int64            `json:"app_updated_user_answers_count"`
	AppCreatedTestSessionsCount      int64            `json:"app_created_test_sessions_count"`
	AppResponsesTotal                map[string]int64 `json:"app_responses_total"`
}

func (c *Snapshot) Hash() string {
	hasher := sha256.New()
	hasher.Write([]byte(strconv.FormatInt(c.AppUsersCreatedCount, 10)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strconv.FormatInt(c.AppNotMessageUpdateCount, 10)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strconv.FormatInt(c.AppGeneratedRecommendationsCount, 10)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strconv.FormatInt(c.AppUpdatedUserAnswersCount, 10)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strconv.FormatInt(c.AppCreatedTestSessionsCount, 10)))
	hasher.Write([]byte{0})
	for path, count := range c.AppResponsesTotal {
		hasher.Write([]byte(fmt.Sprintf("%s - %d", path, count)))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

type Analytics struct {
	log      *slog.Logger
	mu       sync.Mutex
	snapshot Snapshot
}

func (a *Analytics) Snapshot() Snapshot {
	a.mu.Lock()
	m := maps.Clone(a.snapshot.AppResponsesTotal)
	a.mu.Unlock()
	return Snapshot{
		AppUsersCreatedCount:             atomic.LoadInt64(&a.snapshot.AppUsersCreatedCount),
		AppNotMessageUpdateCount:         atomic.LoadInt64(&a.snapshot.AppNotMessageUpdateCount),
		AppGeneratedRecommendationsCount: atomic.LoadInt64(&a.snapshot.AppGeneratedRecommendationsCount),
		AppUpdatedUserAnswersCount:       atomic.LoadInt64(&a.snapshot.AppUpdatedUserAnswersCount),
		AppCreatedTestSessionsCount:      atomic.LoadInt64(&a.snapshot.AppCreatedTestSessionsCount),
		AppResponsesTotal:                m,
	}
}

func (a *Analytics) AppResponsesTotalInc(path string, code int) {
	key := fmt.Sprintf("%s [%d]", path, code)
	a.mu.Lock()
	a.snapshot.AppResponsesTotal[key] = a.snapshot.AppResponsesTotal[key] + 1
	a.mu.Unlock()
}

func (a *Analytics) AppCreatedTestSessionsCountInc() {
	atomic.AddInt64(&a.snapshot.AppCreatedTestSessionsCount, 1)
}

func (a *Analytics) AppUpdatedUserAnswersCountInc() {
	atomic.AddInt64(&a.snapshot.AppUpdatedUserAnswersCount, 1)
}

func (a *Analytics) AppGeneratedRecommendationsCountInc() {
	atomic.AddInt64(&a.snapshot.AppGeneratedRecommendationsCount, 1)
}

func (a *Analytics) AppNotMessageUpdateCountInc() {
	atomic.AddInt64(&a.snapshot.AppNotMessageUpdateCount, 1)
}

func (a *Analytics) AppUsersCreatedCountInc() {
	atomic.AddInt64(&a.snapshot.AppUsersCreatedCount, 1)
}

var _ Metrics = (*Analytics)(nil)

func New(log *slog.Logger) (*Analytics, func()) {
	a := &Analytics{log: log, snapshot: Snapshot{AppResponsesTotal: map[string]int64{}}}
	a.open()
	return a, a.close
}
