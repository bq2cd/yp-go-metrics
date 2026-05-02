package interfaces

type Resettable interface { // not generated
	Reset()
}

type Dummy interface { // not generated
	Dummy()
}

// generate:reset
type EmbeddedInterfaceValue struct {
	Resettable
	Dummy
}

// generate:reset
type InterfaceAsField struct {
	f  Resettable
	fp *Resettable
}
