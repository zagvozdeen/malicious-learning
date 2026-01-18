package analytics

func (a *Analytics) Clone() Metrics {
	return &Analytics{
		log:                              a.log,
		AppUsersCreatedCount:             a.AppUsersCreatedCount,
		AppNotMessageUpdateCount:         a.AppNotMessageUpdateCount,
		AppGeneratedRecommendationsCount: a.AppGeneratedRecommendationsCount,
		AppUpdatedUserAnswersCount:       a.AppUpdatedUserAnswersCount,
		AppCreatedTestSessionsCount:      a.AppCreatedTestSessionsCount,
		AppResponsesTotal:                a.AppResponsesTotal,
	}
}
