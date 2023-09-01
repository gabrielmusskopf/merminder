package template

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"
	"time"
)

var mergeRequestState = []string{"🙂", "🤨", "🙁", "🤒"}

type MergeRequestTemplates struct {
	mergeRequests []*MergeRequestTemplate
}

type MergeRequestTemplate struct {
	MergeRequestTitle                   string
	MergeRequestCount                   int
	MergeRequestStatusE                 string
	MergeRequestOpenTime                string
	MergeRequestTimeSinceLastDiscussion string
	MergeRequestDiscussionResolved      int
	MergeRequestDiscussionCount         int
	MergeRequestURL                     string
}

type MergeRequestInfo struct {
	Title               string
	CreatedAt           time.Time
	TotalDiscussions    int
	OpenDiscussions     int
	TimeOlderDiscussion *time.Time
	URL                 string
}

func ParseMergeRequests(mris []MergeRequestInfo) *MergeRequestTemplates {
	sort.Slice(mris, func(i, j int) bool {
		return mris[i].CreatedAt.Before(mris[j].CreatedAt)
	})

	mtmpl := make([]*MergeRequestTemplate, 0)

	for _, mri := range mris {
		mrt := &MergeRequestTemplate{}

		mrt.MergeRequestCount = len(mris)
		mrt.MergeRequestTitle = mri.Title
		mrt.MergeRequestStatusE = findState(mri)
		mrt.MergeRequestOpenTime = formatTime(time.Since(mri.CreatedAt).Minutes())
		mrt.MergeRequestURL = mri.URL

		if mri.TotalDiscussions > 0 {
			mrt.MergeRequestDiscussionResolved = mri.OpenDiscussions
			mrt.MergeRequestDiscussionCount = mri.TotalDiscussions
		}

		if mri.TimeOlderDiscussion != nil {
			mrt.MergeRequestTimeSinceLastDiscussion = formatTime(time.Since(*mri.TimeOlderDiscussion).Minutes())
		}

		mtmpl = append(mtmpl, mrt)
	}

	return &MergeRequestTemplates{
		mergeRequests: mtmpl,
	}
}

func formatTime(time float64) string {
	opened := fmt.Sprintf("%d minutos", int(time))

	if time >= 60 && time < 60*24 {
		opened = fmt.Sprintf("%d horas", int(time/60))
	} else if time >= 60*24 {
		opened = fmt.Sprintf("%d dias", int(time/(24*60)))
	}
	return opened
}

func daysSince(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}

func findState(mr MergeRequestInfo) string {
	days := daysSince(mr.CreatedAt)

	e := mergeRequestState[3]
	if days >= 0 && days < 3 {
		e = mergeRequestState[0]
	}
	if days >= 2 && days < 4 {
		e = mergeRequestState[1]
	}
	if days >= 4 && days < 6 {
		e = mergeRequestState[2]
	}

	return e
}

func (mrts *MergeRequestTemplates) ParseTemplateFile() (string, error) {
	templFile := "merminder.tmpl"
	template, err := template.New(templFile).ParseFiles(templFile)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, mrts.mergeRequests)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
