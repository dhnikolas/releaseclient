package releaseclient

import (
	"testing"
)

func TestToRepo(t *testing.T) {
	tasks := []Task{
		{
			BranchName: "branch1",
			Services:   []string{"github.com/dhnikolas/test"},
		},
		{
			BranchName: "branch2",
			Services:   []string{"github.com/dhnikolas/test"},
		},
		{
			BranchName: "branch3",
			Services:   []string{"github.com/dhnikolas/test", "github.com/dhnikolas/test2"},
		},
	}
	repos := taskToRepo(tasks)
	if len(repos) != 2 {
		t.Errorf("Repo count error")
	}
	if len(repos[0].Branches) != 3 {
		t.Errorf("Repo %s must have %d branches", repos[0].URL, 3)
	}
	if repos[0].Branches[2].Name != "branch3" {
		t.Errorf("Repo %s must have branch %s", repos[0].URL, "branch3")
	}
}
