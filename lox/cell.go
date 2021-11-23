package lox

//unit processor of world server/game room
type Cell struct {
	*Actor
	Blocks map[int64]Block
}
