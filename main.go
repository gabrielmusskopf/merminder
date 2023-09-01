package main

import (
	"time"

	"github.com/gabrielmusskopf/merminder/config"
	"github.com/gabrielmusskopf/merminder/logger"
	"github.com/gabrielmusskopf/merminder/notify"
	"github.com/gabrielmusskopf/merminder/template"
	"github.com/go-co-op/gocron"
	"github.com/xanzy/go-gitlab"
)

type Merminder struct {
	notifier notify.Notifier
}


func (m *Merminder) fetchMergeReqToApprove(mr *gitlab.MergeRequest, git *gitlab.Client) *template.MergeRequestInfo {
	approval, _, err := git.MergeRequestApprovals.GetConfiguration(mr.ProjectID, mr.IID)
	if err != nil {
		logger.Error(err)
		return nil
	}

	if !approval.Approved {
		comments, _, err := git.Discussions.ListMergeRequestDiscussions(mr.ProjectID, mr.IID, &gitlab.ListMergeRequestDiscussionsOptions{})
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

func (m *Merminder) fetchMergeRequests(git *gitlab.Client) {
	mrsFetched := make(map[int]*gitlab.MergeRequest, 0)
	mrsToApprove := make([]template.MergeRequestInfo, 0)

	if len(config.GetConfig().Observe.Groups) > 0 {
		gOpts := &gitlab.ListGroupMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range config.GetConfig().Observe.Groups {
			mrs, _, err := git.MergeRequests.ListGroupMergeRequests(pid, gOpts)
			if err != nil {
				logger.Error(err)
				return
			}

			for _, mr := range mrs {
				mrsFetched[mr.IID] = mr
				if t := m.fetchMergeReqToApprove(mr, git); t != nil {
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
			mrs, _, err := git.MergeRequests.ListProjectMergeRequests(pid, pOpts)
			if err != nil {
				logger.Error(err)
				return
			}

			for _, mr := range mrs {
				if mrsFetched[mr.IID] != nil {
					continue
				}
				mrsFetched[mr.IID] = mr
				if t := m.fetchMergeReqToApprove(mr, git); t != nil {
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
    m.notifier.Notify(t)
}


func main() {

	config := config.ReadConfig()
    config.LogInfo()

	merminder := &Merminder{
		notifier: notify.NewTeamsNotifier(config.Send.WebhookURL),
	}

	var opt gitlab.ClientOptionFunc
	if !config.DefaultHost() {
		opt = gitlab.WithBaseURL(config.Repository.Host)
	}

	git, err := gitlab.NewClient(config.Repository.Token, opt)
	if err != nil {
		logger.Fatal(err)
	}

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		logger.Fatal(err)
	}

	s := gocron.NewScheduler(location)

	if config.Observe.Every != "" {
		logger.Info("starting merminder with %s update frequency", config.Observe.Every)
		s.Every(config.Observe.Every)

	} else if len(config.Observe.At) != 0 {

		for _, at := range config.Observe.At {
			logger.Info("starting merminder with update scheluded to %s", at)
			s.Every(1).Day().At(at)
		}

	} else {
		logger.Fatals("frequency time is missing. Either configure 'every' or 'at'")
	}

	_, err = s.Do(func() {
		merminder.fetchMergeRequests(git)
	})
	if err != nil {
		logger.Fatal(err)
	}

	s.StartAsync()
	select {}
}
