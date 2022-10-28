package gitclient

import (
	"bytes"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/nomos/go-lokas/log"
	"github.com/pkg/errors"
	"os/exec"
)

type Client struct {
	path string
	repo *git.Repository
}

func NewClient(p string) (*Client, error) {
	ret := &Client{}
	err := ret.SetPath(p)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return ret, nil
}
func (this *Client) SetPath(p string) error {
	var err error
	this.repo, err = git.PlainOpen(p)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) IsClean() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = this.path
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		var exitErr exec.ExitError
		if errors.As(err, &exitErr) {
			return false, errors.Errorf("git status failed: %s", exitErr.Stderr)
		}
		return false, errors.WithStack(err)
	}
	if stdout.Len() > 0 {
		return false, nil
	}
	return true, nil
}

func (this *Client) CurrentBranch() string {
	head, _ := this.repo.Head()
	return head.String()
}

func (this *Client) LastCommit() string {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash()})
	commit, _ := cIter.Next()
	return commit.String()
}

func (this *Client) CurrentTags() []string {
	head, _ := this.repo.Head()
	tags, _ := this.repo.Tags()
	ret := []string{}
	tags.ForEach(func(reference *plumbing.Reference) error {
		if reference.Hash() == head.Hash() {
			ret = append(ret, reference.String())
		}
		return nil
	})
	return ret
}

func (this *Client) GetTagsByCommit(commit string) []string {
	tags, _ := this.repo.Tags()
	ret := []string{}
	tags.ForEach(func(reference *plumbing.Reference) error {
		if reference.Hash().String() == commit {
			ret = append(ret, reference.String())
		}
		return nil
	})
	return ret
}

func (this *Client) LatestTags() []string {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash()})
	tags := []string{}
	for {
		commit, err := cIter.Next()
		if err != nil {
			log.Error(err.Error())
			return tags
		}
		tags = this.GetTagsByCommit(commit.Hash.String())
		if len(tags) > 0 {
			return tags
		}
	}
}
