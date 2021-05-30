package redisclient

type Pair struct {
	Key string
	Value interface{}
}

func CreatePair(k string,v interface{})*Pair {
	return &Pair{
		Key:   k,
		Value: v,
	}
}