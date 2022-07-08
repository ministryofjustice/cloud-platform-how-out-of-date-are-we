package github

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubPRs interface {
	ListPRs(repo string, date date, count int) ([]string, error)
}
type GithubGraphQLClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

type githubPRImpl struct {
	client GithubGraphQLClient
	ctx    context.Context
	token  string
	repo   string
	owner  string
}

func NewGithubPRs(token, repo, owner string) *githubPRImpl {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	oauthClient := oauth2.NewClient(context.Background(), src)

	repository := &githubPRImpl{
		client: githubv4.NewEnterpriseClient("https://api.github.com/graphql", oauthClient),
		ctx:    context.Background(),
		token:  token,
		repo:   repo,
		owner:  owner,
	}

	return repository
}

func (r *githubPRImpl) ListPRs(date date, count int) ([]nodes, error) {

	results, err := r.listPrsForInfra(r.repo, date, count)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type listPrsforInfraQuery struct {
	Search struct {
		Nodes []nodes
	} `graphql:"search(first: $count, query: $searchQuery, type: ISSUE)"`
}

type nodes struct {
	PullRequest struct {
		Title githubv4.String
		Url   githubv4.String
	} `graphql:"... on PullRequest"`
}

type date struct {
	first      time.Time
	last       time.Time
	monthIndex string
}

func (r *githubPRImpl) listPrsForInfra(repo string, date date, count int) ([]nodes, error) {
	query := listPrsforInfraQuery{}

	variables := map[string]interface{}{
		"searchQuery": githubv4.String(fmt.Sprintf(`repo:%s/%s is:pr is:closed merged:%s..%s`, r.owner, repo, date.first.Format("2006-01-02"), date.last.Format("2006-01-02"))),
		"count":       githubv4.Int(count),
	}

	err := r.client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return query.Search.Nodes, nil
}
