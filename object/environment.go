package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	// TODO: consider not having two maps
	c := make(map[string]bool)
	return &Environment{store: s, consts: c, outer: nil}
}

type Environment struct {
	store  map[string]Object
	consts map[string]bool
	outer  *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func (e *Environment) SetConst(name string, val Object) Object {
	e.store[name] = val
	e.consts[name] = true
	return val
}

func (e *Environment) IsConst(name string) bool {
	return e.consts[name]
}
