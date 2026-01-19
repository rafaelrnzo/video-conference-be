package service

import (
	"context"

	"video-conference-be/internal/domain/poll"
)

type PollService struct {
	repo poll.Repository
}

func NewPollService(repo poll.Repository) *PollService {
	return &PollService{repo: repo}
}

func (s *PollService) SavePoll(ctx context.Context, p *poll.Poll) error {
	return s.repo.Create(ctx, p)
}
