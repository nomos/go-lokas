package test

import (
	"github.com/nomos/go-lokas/ecs"
	"testing"
)

func TestRuntime(t *testing.T) {
	_=ecs.CreateECS(1,1,false)
}