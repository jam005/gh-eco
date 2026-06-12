package user

import (
	"testing"

	"github.com/jrnxf/gh-eco/ui/context"
	"github.com/jrnxf/gh-eco/ui/models"
)

func Test_BuildDisplayIsIdempotent(t *testing.T) {
	ctx := &context.ProgramContext{}
	m := NewModel()
	m.UpdateProgramContext(ctx)
	m.User = models.User{
		Login:       "octocat",
		PinnedRepos: []models.Repo{{Id: "r1", Name: "a"}, {Id: "r2", Name: "b"}},
	}
	// BuildGraphDisplay panics on an empty Weeks slice; one empty week keeps
	// buildDisplay from crashing without affecting the widget count assertion.
	m.User.ActivityGraph.Weeks = []models.WeeklyContribution{{}}

	m.buildDisplay()
	first := len(ctx.FocusableWidgets)
	// expect NoFocus + UserDisplay + 2 pinned repos
	if first != 4 {
		t.Fatalf("after first build: got %d widgets, want 4", first)
	}

	m.buildDisplay() // simulates a resize / star-response rebuild
	if got := len(ctx.FocusableWidgets); got != first {
		t.Errorf("after second build: got %d widgets, want %d (rebuild must not duplicate)", got, first)
	}
}
