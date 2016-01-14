package main

import (
	"github.com/zhemao/glisp/interpreter"
)

func defmacroCondValues(env *glisp.Glisp) {
	macro := "(defmac cond-values [ & body] `(cond ~@body))"
	env.EvalString(macro)
}

func NewGlisp() *glisp.Glisp {
	env := glisp.NewGlisp()
	defmacroCondValues(env)
	return env
}
