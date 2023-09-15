package merminder

import (
	"time"

	"github.com/go-co-op/gocron"
	"github.com/xanzy/go-gitlab"
)

type Service struct {
	Notifier Notifier
    //TODO: Support multiple git instances, like GitHub or Bitbucket
	Git      *gitlab.Client
}

func (s *Service) Start() {
	location, err := time.LoadLocation(cfg.Observe.Location)
	if err != nil {
		Fatal(err)
	}

	sched := gocron.NewScheduler(location)

	if cfg.Observe.Every != "" {
		Info("starting merminder with %s update frequency", cfg.Observe.Every)
		sched.Every(cfg.Observe.Every)

	} else if len(cfg.Observe.At) != 0 {
		for _, at := range cfg.Observe.At {
			Info("starting merminder with update scheluded to %s", at)
			sched.Every(1).Day().At(at)
		}

	} else {
		Fatals("frequency time is missing. Either configure 'every' or 'at'")
	}

	_, err = sched.Do(func() {
		s.FetchMergeRequests()
	})
	if err != nil {
		Fatal(err)
	}

	sched.StartAsync()
	select {}
}

func (s *Service) fetchMergeReqToApprove(mr *gitlab.MergeRequest) *MergeRequestInfo {
	approval, _, err := s.Git.MergeRequestApprovals.GetConfiguration(mr.ProjectID, mr.IID)
	if err != nil {
		Error(err)
		return nil
	}

	if !approval.Approved {
		comments, _, err := s.Git.Discussions.ListMergeRequestDiscussions(mr.ProjectID, mr.IID, &gitlab.ListMergeRequestDiscussionsOptions{})
		if err != nil {
			Error(err)
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

		return &MergeRequestInfo{
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
	mrsToApprove := make([]MergeRequestInfo, 0)

	if len(GetConfig().Observe.Groups) > 0 {
		gOpts := &gitlab.ListGroupMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range GetConfig().Observe.Groups {
			mrs, _, err := s.Git.MergeRequests.ListGroupMergeRequests(pid, gOpts)
			if err != nil {
				Error(err)
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

	if len(GetConfig().Observe.Projects) > 0 {
		pOpts := &gitlab.ListProjectMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range GetConfig().Observe.Projects {
			mrs, _, err := s.Git.MergeRequests.ListProjectMergeRequests(pid, pOpts)
			if err != nil {
				Error(err)
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

	Info("total MRs fetched according to config: %d\n", len(mrsToApprove))

	t, err := ParseMergeRequests(mrsToApprove).ParseTemplateFile()
	if err != nil {
		Error(err)
		return
	}
	if GetConfig().NotificationEnabled() {
		if err := s.Notifier.Notify(t); err != nil {
			Error(err)
		}
	}
}
