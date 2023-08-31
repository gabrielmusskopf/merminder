package main

import (
	"os"
	"time"

	"github.com/gabrielmusskopf/merminder/logger"
	"github.com/gabrielmusskopf/merminder/notify"
	"github.com/go-co-op/gocron"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v3"
)

type Merminder struct {
	config   *Config
	notifier notify.Notifier
}

type Config struct {
	Repository struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	}
	Send struct {
		WebhookURL string `yaml:"webhookUrl"`
	}
	Observe struct {
		Groups   []int    `yaml:",flow"`
		Projects []int    `yaml:",flow"`
		Every    string   `yaml:"every"`
		At       []string `yaml:",flow"`
	}
}

func (m *Merminder) fetchMergeReqToApprove(mr *gitlab.MergeRequest, git *gitlab.Client) *notify.MergeRequestInfo {
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

		return &notify.MergeRequestInfo{
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
	mrsToApprove := make([]notify.MergeRequestInfo, 0)

	if len(m.config.Observe.Groups) > 0 {
		gOpts := &gitlab.ListGroupMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range m.config.Observe.Groups {
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

	if len(m.config.Observe.Projects) > 0 {
		pOpts := &gitlab.ListProjectMergeRequestsOptions{
			State: gitlab.String("opened"),
		}
		for _, pid := range m.config.Observe.Projects {
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
	m.notifier.Notify(mrsToApprove)
}

func readConfig() *Config {
	f, err := os.Open(".merminder.yml")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	config := &Config{}

	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&config); err != nil {
		logger.Fatal(err)
	}

	if config.Repository.Token == "" {
		logger.Fatals("token is missing")
	}

    if config.Observe.Every != "" && len(config.Observe.At) != 0 {
        logger.Warning("cannot use 'observe.at' and 'obser.every' at the same time")
        logger.Warning("only 'observe.every' will be considered")
        config.Observe.At = make([]string, 0)
    } else {
        logger.Fatals("at least one observe frequency must be set: 'every' or 'at'")
    }

	return config
}

func (c *Config) LogInfo() {
    logger.Info("repository url: %s", c.Repository.Host)
    logger.Info("webhook url: %s", c.Send.WebhookURL)
    logger.Info("observed groups: %v", c.Observe.Groups)
    logger.Info("observed projects: %v", c.Observe.Projects)
    if c.Observe.Every != "" {
        logger.Info("every: %s", c.Observe.Every)
    } else if len(c.Observe.At) != 0 {
        logger.Info("at: %s", c.Observe.At)
    }
}

func (c *Config) DefaultHost() bool {
	return c.Repository.Host == ""
}

func main() {

	config := readConfig()
    config.LogInfo()

	merminder := &Merminder{
		config:   config,
		notifier: notify.NewTeamsNotifier(config.Send.WebhookURL),
	}

	var opt gitlab.ClientOptionFunc
	if !merminder.config.DefaultHost() {
		opt = gitlab.WithBaseURL(merminder.config.Repository.Host)
	}

	git, err := gitlab.NewClient(merminder.config.Repository.Token, opt)
	if err != nil {
		logger.Fatal(err)
	}

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		logger.Fatal(err)
	}

	s := gocron.NewScheduler(location)

	if merminder.config.Observe.Every != "" {
		logger.Info("starting merminder with %s update frequency", merminder.config.Observe.Every)
		s.Every(merminder.config.Observe.Every)

	} else if len(merminder.config.Observe.At) != 0 {

		for _, at := range merminder.config.Observe.At {
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
