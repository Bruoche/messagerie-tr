package arg

import (
	"flag"
)

type Key[T any] int

type SupportedType[T any] interface {
	Bind(name, usage string)
	Default() T
}

type bindable[T any] struct {
	ptr *T
	def T
}

func FromDefault[T any](val T) bindable[T] {
	return bindable[T]{
		ptr: &val,
		def: val,
	}
}

func (val bindable[T]) Default() T {
	return val.def
}

type Flag[T any] struct {
	bind  SupportedType[T]
	name  string
	usage string
}

var _ Flaggable = Flag[int]{}

func Make[T any](name string, bind SupportedType[T], usage string) Flaggable {
	return Flag[T]{
		bind:  bind,
		name:  name,
		usage: usage,
	}
}

func (f Flag[T]) Bind() {
	f.bind.Bind(f.name, f.usage)
}

type Flaggable interface {
	Bind()
}

func Parse(flags []Flaggable) {
	for _, flag := range flags {
		flag.Bind()
	}
	flag.Parse()
}

type Int struct {
	bindable[int]
}

var _ SupportedType[int] = &Int{}

func (val Int) Bind(name, usage string) {
	flag.IntVar(val.ptr, name, val.Default(), usage)
}

type String struct {
	bindable[string]
}

var _ SupportedType[string] = &String{}

func (val String) Bind(name, usage string) {
	flag.StringVar(val.ptr, name, val.Default(), usage)
}

type Bool struct {
	bindable[bool]
}

var _ SupportedType[bool] = &Bool{}

func (val Bool) Bind(name, usage string) {
	flag.BoolVar(val.ptr, name, val.Default(), usage)
}
