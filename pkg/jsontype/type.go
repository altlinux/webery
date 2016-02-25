package jsontype

type Export interface {
	IsDefined() bool
	String() string
}
