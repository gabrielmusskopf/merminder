package service

import (
	"time"

	"github.com/gabrielmusskopf/merminder/config"
	"github.com/gabrielmusskopf/merminder/logger"
	"github.com/gabrielmusskopf/merminder/notify"
	"github.com/gabrielmusskopf/merminder/template"
	"github.com/xanzy/go-gitlab"
)

type Service struct {
	Notifier notify.Notifier
    //TODO: Support multiple git instances, like GitHub or Bitbucket
	Git      *gitlab.Client
}

func (s *Service) fetchMergeReqToApprove(mr *gitlab.MergeRequest) *template.MergeRequestInfo {
	approval, _, err := s.Git.MergeRequestApprovals.GetConfiguration(mr.ProjectID, mr.IID)
	if err != nil {
		logger.Error(err)
		return nil
	}

	if !approval.Approved {
		comments, _, err := s.Git.Discussions.ListMergeRequestDiscussions(mr.ProjectID, mr.IID, &gitlab.ListMergeRequestDiscussionsOptions{})
		if err != nil {
			logger.Error(err)
			return nil
		}

		var discussions int
		var discussionsOpen int
		dateOlderDiscussion := &time.Time{}
		for _, c := range comments {
			for _, note := range c.Notes {
				if note.Type == "DiffNote" {
					discussions += 1

					if note.Resolved {
						discussionsOpen += 1
					}
				}
				if note.CreatedAt.After(*dateOlderDiscussion) {
					dateOlderDiscussion = note.CreatedAt
				}
			}
		}

		if discussions <= 0 {
			dateOlderDiscussion = nil
		}

		return &template.MergeRequestInfo{
			Title:               mr.Title,
			CreatedAt:           *mr.CreatedAt,
			TotalDiscussions:    discussions,
			OpenDiscussions:     discussionsOpen,
			TimeOlderDiscussion: dateOlderDiscussion,
			URL:                 mr.WebURL,
		}
	}
	return nil
}

func (s *Service) FetchMergeRequests() {
	mrsFetched := make(map[int]*gitlab.MergeRequest, 0)
	mrsToApprove := make([]template.MergeRequestInfo, 0)

	if len(config.GetConfig().Observe.Groups) > 0 {
		gOpts := &gitlab.ListGroupMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range config.GetConfig().Observe.Groups {
			mrs, _, err := s.Git.MergeRequests.ListGroupMergeRequests(pid, gOpts)
			if err != nil {
				logger.Error(err)
				return
			}

			for _, mr := range mrs {
				mrsFetched[mr.IID] = mr
				if t := s.fetchMergeReqToApprove(mr); t != nil {
					mrsToApprove = append(mrsToApprove, *t)
				}
			}
		}
	}

	if len(config.GetConfig().Observe.Projects) > 0 {
		pOpts := &gitlab.ListProjectMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range config.GetConfig().Observe.Projects {
			mrs, _, err := s.Git.MergeRequests.ListProjectMergeRequests(pid, pOpts)
			if err != nil {
				logger.Error(err)
				return
			}

			for _, mr := range mrs {
				if mrsFetched[mr.IID] != nil {
					continue
				}
				mrsFetched[mr.IID] = mr
				if t := s.fetchMergeReqToApprove(mr); t != nil {
					mrsToApprove = append(mrsToApprove, *t)
				}
			}
		}
	}

	logger.Info("total MRs fetched according to config: %d\n", len(mrsToApprove))

	t, err := template.ParseMergeRequests(mrsToApprove).ParseTemplateFile()
	if err != nil {
		logger.Error(err)
		return
	}
	if config.GetConfig().NotificationEnabled() {
		if err := s.Notifier.Notify(t); err != nil {
			logger.Error(err)
		}
	}
}
