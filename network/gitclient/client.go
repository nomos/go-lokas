package gitclient

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/slice"
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

func (this *Client) IsClean() bool {
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

func (this *Client) CurrentTags() []*object.Tag {
	head, _ := this.repo.Head()
	tags, _ := this.repo.TagObjects()
	ret := []*object.Tag{}
	tags.ForEach(func(tag *object.Tag) error {
		if tag.Target.String() == head.Hash().String() {
			ret = append(ret, tag)
		}
		return nil
	})
	return ret
}

func (this *Client) GetCommitByTagName(s string) *object.Commit {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash()})
	tags := map[string]string{}
	tagIter, _ := this.repo.TagObjects()
	tagIter.ForEach(func(tag *object.Tag) error {
		tags[tag.Name] = tag.Target.String()
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
	tags, _ := this.repo.TagObjects()
	ret := []string{}
	tags.ForEach(func(tag *object.Tag) error {
		if tag.Target.String() == commit {
			log.Infof(tag.Name)
			ret = append(ret, tag.Name)
		}
		return nil
	})
	return ret
}

func (this *Client) LatestTag() *object.Tag {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash(), Order: git.LogOrderCommitterTime})
	tags := map[string][]*object.Tag{}
	tagIter, _ := this.repo.TagObjects()
	tagIter.ForEach(func(tag *object.Tag) error {
		if tags[tag.Target.String()] == nil {
			tags[tag.Target.String()] = []*object.Tag{}
		}
		tags[tag.Target.String()] = append(tags[tag.Target.String()], tag)
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

func (this *Client) LatestTags() []*object.Tag {
	head, _ := this.repo.Head()
	cIter, _ := this.repo.Log(&git.LogOptions{From: head.Hash(), Order: git.LogOrderCommitterTime})
	tags := map[string][]*object.Tag{}
	tagIter, _ := this.repo.TagObjects()
	tagIter.ForEach(func(tag *object.Tag) error {
		if tags[tag.Target.String()] == nil {
			tags[tag.Target.String()] = []*object.Tag{}
		}
		tags[tag.Target.String()] = append(tags[tag.Target.String()], tag)
		return nil
	})
	ret := []*object.Tag{}
	cIter.ForEach(func(commit *object.Commit) error {
		if t, ok := tags[commit.Hash.String()]; ok {
			ret = slice.Concat(ret, t)
		}
		return nil
	})
	return ret
}
