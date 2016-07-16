package highlight

type Emitter interface {
	Emit(t Token) (int, error)
}
