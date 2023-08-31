package notify

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gabrielmusskopf/merminder/logger"
)

type MSTeamsNotifier struct {
	Url string
}

func NewTeamsNotifier(url string) *MSTeamsNotifier {
	return &MSTeamsNotifier{
		Url: url,
	}
}

var mergeRequestState = []string{"🙂","🤨","🙁","🤒"}

func daysSince(t time.Time) int {
    return int(time.Since(t).Hours() / 24)
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

//TODO: Mover formatação para arquivo template

func (ms MSTeamsNotifier) Notify(mris []MergeRequestInfo) {

	sort.Slice(mris, func(i, j int) bool {
		return mris[i].CreatedAt.Before(mris[j].CreatedAt)
	})

    var sectionsBuilder strings.Builder

    sectionsBuilder.WriteString(fmt.Sprintf(`{
        text: "<blockquote><h1><strong>Quantidade de merge requests: %d</strong></h1></blockquote><br/>"
    },`,len(mris)))

	for i, mri := range mris {

        var builder strings.Builder

        builder.WriteString(fmt.Sprintf(`<h1><strong>%s</strong></h1><h2><strong>Status:</strong> %s</h2>`, mri.Title, findState(mri)))

        openedSince := formatTime(time.Since(mri.CreatedAt).Minutes())
        builder.WriteString(fmt.Sprintf("<h2><strong>Aberto há:</strong> %s</h2>", openedSince))

        if mri.TimeOlderDiscussion != nil {
            unresolvedComment := formatTime(time.Since(*mri.TimeOlderDiscussion).Minutes())
            builder.WriteString(fmt.Sprintf("<h2><strong>Comentário mais antigo aberto há:</strong> %s</h2>", unresolvedComment))
        }

        if mri.TotalDiscussions > 0 {
            builder.WriteString(fmt.Sprintf("<h2><strong>Comentários resolvidos:</strong> %d/%d</h2>", mri.OpenDiscussions, mri.TotalDiscussions))
        }

        builder.WriteString(fmt.Sprintf("<h2><strong>URL:</strong> <a href=\\\"%s\\\">%s</a></h2>", mri.URL, mri.URL))

		if i != (len(mris) - 1) {
            builder.WriteString("<br/>")
		}

        sectionsBuilder.WriteString(fmt.Sprintf("{text: \"%s\"},", builder.String()))
	}

    body := fmt.Sprintf(`{
        "type": "message",
        "attachments": [
        {
            "contentType": "application/vnd.microsoft.teams.card.o365connector",
            "content": {
                "@type": "MessageCard",
                "@context": "https://schema.org/extensions",
                "summary": "Summary",
                "title": "Merge Requests",
                "sections": [%s]
            }
        }
        ]
    }`, sectionsBuilder.String())

	if err := Send(ms.Url, []byte(body)); err != nil {
		logger.Error(err)
	}

}
