package lokas

func Get[T IComponent](entity IEntity) T {
	var t T
	id, _ := t.GetId()
	return entity.Get(id).(T)
}

func Remove[T IComponent](entity IEntity) T {
	var t T
	id, _ := t.GetId()
	return entity.Remove(id).(T)
}
