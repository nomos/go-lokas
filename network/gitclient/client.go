package gitclient

import (
	"bytes"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/slice"
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
	this.path = p
	this.repo, err = git.PlainOpen(p)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) IsClean() bool {
	return this.isCleanShell()
}

func (this *Client) isClean() bool {
	tree, err := this.repo.Worktree()
	if err != nil {
		log.Error(err.Error())
		return false
	}
	status, err := tree.Status()
	if err != nil {
		log.Error(err.Error())
		return false
	}
	return status.IsClean()
}

func (this *Client) isCleanShell() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = this.path
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		var exitErr exec.ExitError
		if errors.As(err, &exitErr) {
			log.Errorf("git status failed: %s", exitErr.Stderr)
			return false
		}
		log.Errorf("git status failed: %s", errors.WithStack(err))
		return false
	}
	if stdout.Len() > 0 {
		return false
	}
	return true
}

func (this *Client) CurrentBranch() string {
	head, _ := this.repo.Head()
	return head.String()
}

func (this *Client) LastCommit() *object.Commit {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash()})
	commit, _ := cIter.Next()
	return commit
}

func (this *Client) CurrentTags() []*Tag {
	head, _ := this.repo.Head()
	tags, _ := this.repo.Tags()
	ret := []*Tag{}
	tags.ForEach(func(reference *plumbing.Reference) error {
		tagName := reference.Name().Short()
		hash := reference.Hash().String()
		if hash == head.Hash().String() {
			ret = append(ret, &Tag{
				Name: tagName,
				Hash: hash,
			})
		}
		return nil
	})
	return ret
}

func (this *Client) GetCommitByTagName(s string) *object.Commit {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash()})
	tags := map[string]string{}
	tagIter, _ := this.repo.Tags()
	tagIter.ForEach(func(reference *plumbing.Reference) error {
		tagName := reference.Name().Short()
		hash := reference.Hash().String()
		tags[tagName] = hash
		return nil
	})
	hashStr := tags[s]
	if hashStr == "" {
		return nil
	}
	var ret *object.Commit
	cIter.ForEach(func(commit *object.Commit) error {
		if commit.Hash.String() == hashStr {
			ret = commit
			return storer.ErrStop
		}
		return nil
	})
	return ret
}

func (this *Client) GetTagsByCommit(commit string) []string {
	tags, _ := this.repo.Tags()
	ret := []string{}
	tags.ForEach(func(reference *plumbing.Reference) error {
		tagName := reference.Name().Short()
		if reference.Hash().String() == commit {
			ret = append(ret, tagName)
		}
		return nil
	})
	return ret
}

type Tag struct {
	Name string
	Hash string
}

func (this *Client) LatestTag() *Tag {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash(), Order: git.LogOrderCommitterTime})
	tags := map[string][]*Tag{}
	tagIter, _ := this.repo.Tags()

	tagIter.ForEach(func(reference *plumbing.Reference) error {
		tagName := reference.Name().Short()
		hash := reference.Hash().String()
		if tags[hash] == nil {
			tags[hash] = []*Tag{}
		}
		tags[hash] = append(tags[hash], &Tag{
			Name: tagName,
			Hash: hash,
		})
		return nil
	})
	for {
		item, err := cIter.Next()
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		if t, ok := tags[item.Hash.String()]; ok {
			return t[0]
		}
	}
	return nil
}

func (this *Client) LatestTags() []*Tag {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash()})
	tags := map[string][]*Tag{}
	tagIter, _ := this.repo.Tags()
	tagIter.ForEach(func(reference *plumbing.Reference) error {
		tagName := reference.Name().Short()
		hash := reference.Hash().String()
		if tags[hash] == nil {
			tags[hash] = []*Tag{}
		}
		tags[hash] = append(tags[hash], &Tag{
			Name: tagName,
			Hash: hash,
		})
		return nil
	})
	ret := []*Tag{}
	cIter.ForEach(func(commit *object.Commit) error {
		if t, ok := tags[commit.Hash.String()]; ok {
			ret = slice.Concat(ret, t)
		}
		return nil
	})
	return ret
}
