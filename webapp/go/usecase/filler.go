package usecase

// orderedBy は filled (Filler の Fill の結果) の値を、models と同じ順序 (重複も含む) で並べて返す。
// id は model から filled のキー (ID) を取り出す。
func orderedBy[M any, K comparable, V any](models []M, id func(M) K, filled map[K]V) []V {
	values := make([]V, len(models))
	for i, model := range models {
		values[i] = filled[id(model)]
	}
	return values
}
