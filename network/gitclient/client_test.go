package gitclient

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/nomos/go-lokas/log"
	"strings"
	"testing"
)

func TestGitClient(t *testing.T) {

	repo, _ := git.PlainOpen("/Users/wqs/tingzhou/projecta/Data")
	d, _ := repo.Tags()
	d.ForEach(func(reference *plumbing.Reference) error {
		log.Infof(strings.ReplaceAll(reference.Name().String(), "refs/tags/", ""), reference.Hash().String())
		return nil
	})

}
