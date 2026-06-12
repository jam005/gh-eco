package utils

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/ui/models"
)

func TruncateText(str string, max int) string {
	if max <= 0 {
		return ""
	}

	lastSpaceIdx := -1
	count := 0
	hardCutIdx := -1 // byte index of the first rune past max; -1 until exceeded
	for i, r := range str {
		if unicode.IsSpace(r) {
			if hardCutIdx != -1 {
				// already past max with no earlier space: cut at this one
				return str[:i] + "..."
			}
			lastSpaceIdx = i
		}
		count++
		if count > max && hardCutIdx == -1 {
			if lastSpaceIdx != -1 {
				return str[:lastSpaceIdx] + "..."
			}
			hardCutIdx = i
		}
	}
	if hardCutIdx != -1 {
		// longer than max with no spaces at all: hard cut at max runes
		return str[:hardCutIdx] + "..."
	}
	return str
}

func BrowserOpen(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Println(err)
	}
}

func GetNewLines(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("\n", n)
}

func MapGetUserQueryToDisplayUser(query queries.GetUserQuery) models.User {
	qu := query.User
	du := models.User{
		Login:             qu.Login,
		Name:              qu.Name,
		Location:          qu.Location,
		Url:               qu.Url,
		Bio:               qu.Bio,
		TwitterUsername:   qu.TwitterUsername,
		WebsiteUrl:        qu.WebsiteUrl,
		FollowersCount:    qu.Followers.TotalCount,
		FollowingCount:    qu.Following.TotalCount,
		IsViewer:          qu.IsViewer,
		IsFollowingViewer: qu.IsFollowingViewer,
		ViewerIsFollowing: qu.ViewerIsFollowing,
	}

	du.ActivityGraph.ContributionsCount = qu.ContributionsCollection.ContributionCalendar.TotalContributions

	for _, week := range qu.ContributionsCollection.ContributionCalendar.Weeks {
		du.ActivityGraph.Weeks = append(du.ActivityGraph.Weeks, week)
	}

	for _, node := range qu.PinnedItems.Nodes {
		r := node.Repository
		du.PinnedRepos = append(du.PinnedRepos, models.Repo{
			Id:               r.Id,
			Name:             r.Name,
			Description:      r.Description,
			StarsCount:       r.StargazerCount,
			ViewerHasStarred: r.ViewerHasStarred,
			Owner: struct{ Login string }{
				Login: r.Owner.Login,
			},
			Url:             r.Url,
			PrimaryLanguage: r.PrimaryLanguage,
		})
	}

	return du
}

func MaxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
