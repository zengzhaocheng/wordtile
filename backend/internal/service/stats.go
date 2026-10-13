package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/repository"
)

// StatsService 数据统计：活跃度砖块图、题库用时/错误率演进、闯关战绩。
type StatsService struct {
	repo *repository.DB
}

func NewStatsService(repo *repository.DB) *StatsService { return &StatsService{repo: repo} }

// ContributionDay 活跃度砖块图单日数据。
type ContributionDay struct {
	Date      string  `json:"date"`
	Sessions  int64   `json:"sessions"`
	Duration  int64   `json:"duration_ms"`
	Accuracy  float64 `json:"accuracy"`
}

func (s *StatsService) Contribution(userID uuid.UUID) ([]ContributionDay, error) {
	rows, err := s.repo.ContributionByDay(userID)
	if err != nil {
		return nil, err
	}
	out := make([]ContributionDay, 0, len(rows))
	for _, r := range rows {
		var acc float64
		if r.Correct+r.Wrong > 0 {
			acc = float64(r.Correct) / float64(r.Correct+r.Wrong)
		}
		out = append(out, ContributionDay{Date: r.Date, Sessions: r.Sessions, Duration: r.Duration, Accuracy: acc})
	}
	return out, nil
}

// ProgressPoint 某题库一局的时间序列点。
type ProgressPoint struct {
	CompletedAt time.Time `json:"completed_at"`
	DurationMS  int64     `json:"duration_ms"`
	Correct     int       `json:"correct_count"`
	Wrong       int       `json:"wrong_count"`
	ErrorRate   float64   `json:"error_rate"` // 错误匹配数 / 正确匹配数
}

func (s *StatsService) DeckProgress(userID, deckID uuid.UUID) ([]ProgressPoint, error) {
	sessions, err := s.repo.ListFinishedSessions(userID)
	if err != nil {
		return nil, err
	}
	out := []ProgressPoint{}
	for _, sess := range sessions {
		if sess.DeckID != deckID {
			continue
		}
		var errRate float64
		if sess.CorrectCount > 0 {
			errRate = float64(sess.WrongCount) / float64(sess.CorrectCount)
		} else if sess.WrongCount > 0 {
			errRate = 1
		}
		out = append(out, ProgressPoint{
			CompletedAt: *sess.CompletedAt, DurationMS: sess.DurationMS,
			Correct: sess.CorrectCount, Wrong: sess.WrongCount, ErrorRate: errRate,
		})
	}
	return out, nil
}

// CampaignStat 闯关模式关卡战绩。
type CampaignStat struct {
	Level     int   `json:"level"`
	Plays     int64 `json:"plays"`
	BestMS    int64 `json:"best_ms"`
	BestWrong int64 `json:"best_wrong"`
}

func (s *StatsService) Campaign(userID uuid.UUID) ([]CampaignStat, error) {
	rows, err := s.repo.CampaignByLevel(userID)
	if err != nil {
		return nil, err
	}
	out := make([]CampaignStat, 0, len(rows))
	for _, r := range rows {
		out = append(out, CampaignStat(r))
	}
	return out, nil
}
