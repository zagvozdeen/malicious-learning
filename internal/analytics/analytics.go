package analytics

import (
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"sync/atomic"
)

type Metrics interface {
	Clone() Metrics

	AppUsersCreatedCountInc()
	AppNotMessageUpdateCountInc()
	AppGeneratedRecommendationsCountInc()
	AppUpdatedUserAnswersCountInc()
	AppCreatedTestSessionsCountInc()
	AppResponsesTotalInc(path string, code int)

	GetAppUsersCreatedCount() int64
	GetAppNotMessageUpdateCount() int64
	GetAppGeneratedRecommendationsCount() int64
	GetAppUpdatedUserAnswersCount() int64
	GetAppCreatedTestSessionsCount() int64
	GetAppResponsesTotal() map[string]int64
}

type Analytics struct {
	log *slog.Logger

	AppUsersCreatedCount             int64 `json:"app_users_created_count"`
	AppNotMessageUpdateCount         int64 `json:"app_not_message_update_count"`
	AppGeneratedRecommendationsCount int64 `json:"app_generated_recommendations_count"`
	AppUpdatedUserAnswersCount       int64 `json:"app_updated_user_answers_count"`
	AppCreatedTestSessionsCount      int64 `json:"app_created_test_sessions_count"`

	AppResponsesTotal   map[string]int64 `json:"app_responses_total"`
	appResponsesTotalMu sync.Mutex
}

func (a *Analytics) Clone() Metrics {
	return &Analytics{
		log:                              a.log,
		AppUsersCreatedCount:             a.GetAppUsersCreatedCount(),
		AppNotMessageUpdateCount:         a.GetAppNotMessageUpdateCount(),
		AppGeneratedRecommendationsCount: a.GetAppGeneratedRecommendationsCount(),
		AppUpdatedUserAnswersCount:       a.GetAppUpdatedUserAnswersCount(),
		AppCreatedTestSessionsCount:      a.GetAppCreatedTestSessionsCount(),
		AppResponsesTotal:                a.GetAppResponsesTotal(),
	}
}

func (a *Analytics) AppResponsesTotalInc(path string, code int) {
	key := fmt.Sprintf("%s [%d]", path, code)
	a.appResponsesTotalMu.Lock()
	a.AppResponsesTotal[key] = a.AppResponsesTotal[key] + 1
	a.appResponsesTotalMu.Unlock()
}

func (a *Analytics) GetAppResponsesTotal() map[string]int64 {
	a.appResponsesTotalMu.Lock()
	defer a.appResponsesTotalMu.Unlock()
	return maps.Clone(a.AppResponsesTotal)
}

func (a *Analytics) AppCreatedTestSessionsCountInc() {
	atomic.AddInt64(&a.AppCreatedTestSessionsCount, 1)
}

func (a *Analytics) GetAppCreatedTestSessionsCount() int64 {
	return atomic.LoadInt64(&a.AppCreatedTestSessionsCount)
}

func (a *Analytics) AppUpdatedUserAnswersCountInc() {
	atomic.AddInt64(&a.AppUpdatedUserAnswersCount, 1)
}

func (a *Analytics) GetAppUpdatedUserAnswersCount() int64 {
	return atomic.LoadInt64(&a.AppUpdatedUserAnswersCount)
}

func (a *Analytics) AppGeneratedRecommendationsCountInc() {
	atomic.AddInt64(&a.AppGeneratedRecommendationsCount, 1)
}

func (a *Analytics) GetAppGeneratedRecommendationsCount() int64 {
	return atomic.LoadInt64(&a.AppGeneratedRecommendationsCount)
}

func (a *Analytics) AppNotMessageUpdateCountInc() {
	atomic.AddInt64(&a.AppNotMessageUpdateCount, 1)
}

func (a *Analytics) GetAppNotMessageUpdateCount() int64 {
	return atomic.LoadInt64(&a.AppNotMessageUpdateCount)
}

func (a *Analytics) AppUsersCreatedCountInc() {
	atomic.AddInt64(&a.AppUsersCreatedCount, 1)
}

func (a *Analytics) GetAppUsersCreatedCount() int64 {
	return atomic.LoadInt64(&a.AppUsersCreatedCount)
}

var _ Metrics = (*Analytics)(nil)

func New(log *slog.Logger) (*Analytics, func()) {
	a := &Analytics{log: log, AppResponsesTotal: map[string]int64{}}
	a.open()
	return a, a.close
}
