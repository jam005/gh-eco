package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jrnxf/gh-eco/api/github/mutations"
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/models"
	"github.com/jrnxf/gh-eco/utils"
	ghv4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const GH_ECO_REPO_ID string = "R_kgDOHVAImQ"

var (
	clientInstance *ghv4.Client
	clientErr      error
	once           sync.Once
)

// GetClient initializes a GitHub GraphQL client instance with a token obtained from GitHub CLI.
func GetClient() (*ghv4.Client, error) {
	once.Do(func() {
		output, _, err := gh.Exec("auth", "token")
		if err != nil {
			clientErr = fmt.Errorf("unable to retrieve gh auth token: %w", err)
			return
		}

		token := strings.TrimSpace(output.String())
		if token == "" {
			clientErr = fmt.Errorf("empty gh auth token")
			return
		}

		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient := oauth2.NewClient(context.Background(), src)
		clientInstance = ghv4.NewClient(httpClient)
	})

	return clientInstance, clientErr
}

func GetUser(login string) tea.Cmd {
	return func() tea.Msg {
		client, err := GetClient()
		if err != nil {
			return commands.GetUserResponse{Err: err}
		}

		var query queries.GetUserQuery

		variables := map[string]interface{}{
			"login": ghv4.String(login),
			"first": ghv4.Int(6),
		}

		err = client.Query(context.Background(), &query, variables)
		if err != nil {
			log.Println(err)
			return commands.GetUserResponse{Err: err}
		}

		return commands.GetUserResponse{User: utils.MapGetUserQueryToDisplayUser(query)}
	}
}

func GetReadme(name string, owner string) tea.Cmd {
	return func() tea.Msg {
		client, err := api.DefaultRESTClient()
		if err != nil {
			return commands.GetReadmeResponse{Err: err}
		}

		var response struct {
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}

		path := fmt.Sprintf("repos/%s/%s/readme", url.PathEscape(owner), url.PathEscape(name))
		if err := client.Get(path, &response); err != nil {
			log.Println(err)
			return commands.GetReadmeResponse{Err: err}
		}

		text, err := decodeReadmeContent(response.Content, response.Encoding)
		if err != nil {
			log.Println(err)
			return commands.GetReadmeResponse{Err: err}
		}

		return commands.GetReadmeResponse{Readme: models.Blob{Text: text}}
	}
}

// decodeReadmeContent decodes the content field of the REST readme
// response. GitHub returns base64 with embedded newlines, which the
// strict std decoder rejects, so strip them first.
func decodeReadmeContent(content, encoding string) (string, error) {
	if encoding != "base64" {
		return "", fmt.Errorf("unexpected readme encoding %q", encoding)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func StarStarrable(starrableId string) tea.Cmd {
	return func() tea.Msg {
		client, err := GetClient()
		if err != nil {
			return commands.StarStarrableResponse{Err: err}
		}

		var mutation mutations.AddStarMutation

		input := ghv4.AddStarInput{
			StarrableID: ghv4.ID(starrableId),
		}

		err = client.Mutate(context.Background(), &mutation, input, nil)
		if err != nil {
			log.Println(err)
			return commands.StarStarrableResponse{Err: err}
		}

		return commands.StarStarrableResponse{Starrable: mutation.AddStar.Starrable}
	}
}

func RemoveStarStarrable(starrableId string) tea.Cmd {
	return func() tea.Msg {
		client, err := GetClient()
		if err != nil {
			return commands.RemoveStarStarrableResponse{Err: err}
		}

		var mutation mutations.RemoveStarMutation

		input := ghv4.RemoveStarInput{
			StarrableID: ghv4.ID(starrableId),
		}

		err = client.Mutate(context.Background(), &mutation, input, nil)
		if err != nil {
			log.Println(err)
			return commands.RemoveStarStarrableResponse{Err: err}
		}

		return commands.RemoveStarStarrableResponse{Starrable: mutation.RemoveStar.Starrable}
	}
}
