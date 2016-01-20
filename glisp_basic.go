package main

import (
	"github.com/hashicorp/go-version"
	"github.com/zhemao/glisp/interpreter"
	"regexp"
	"strings"
	"unicode/utf8"
)

func defmacroCondValues(env *glisp.Glisp) {
	macro := "(defmac cond-values [ & body] `(cond ~@body))"
	env.EvalString(macro)
}

// version-cmp
func versionCompareFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var vleft, vright string

	switch t := args[0].(type) {
	case glisp.SexpStr:
		vleft = string(t)
	default:
		//return glisp.SexpNull, errors.New("version argument must be string")
		return glisp.SexpNull, nil
	}

	switch t := args[1].(type) {
	case glisp.SexpStr:
		vright = string(t)
	default:
		//return glisp.SexpNull, errors.New("version argument must be string")
		return glisp.SexpNull, nil
	}

	v1, err := version.NewVersion(vleft)
	if err != nil {
		return glisp.SexpNull, nil
		//return glisp.SexpNull, errors.New("version format error")
	}
	v2, err := version.NewVersion(vright)
	if err != nil {
		return glisp.SexpNull, nil
		//return glisp.SexpNull, errors.New("version format error")
	}
	res := v1.Compare(v2)
	return glisp.SexpInt(res), nil
}

func defVersionCompareFunctions(env *glisp.Glisp) {
	env.AddFunction("version-cmp", versionCompareFunction)
	shortcuts := `
         (defn ver= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (= ret 0) false)))
         (defn ver> [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (> ret 0) false)))
         (defn ver< [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (< ret 0) false)))
         (defn ver>= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (>= ret 0) false)))
         (defn ver<= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (<= ret 0) false)))
         (defn ver!= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (not= ret 0) false)))
    `
	env.EvalString(shortcuts)
}

// string functions

func stringWildcardMatchFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var pattern, target string
	var regexp_pattern string = ""

	switch t := args[0].(type) {
	case glisp.SexpStr:
		pattern = string(t)
	default:
		//return glisp.SexpNull, errors.New("wildcard parttern must be string")
		return glisp.SexpNull, nil
	}

	switch t := args[1].(type) {
	case glisp.SexpStr:
		target = string(t)
	default:
		//return glisp.SexpNull, errors.New("arg1 must be string")
		return glisp.SexpNull, nil
	}

	for i, w, l := 0, 0, 0; i < len(pattern); i += w {
		runeValue, width := utf8.DecodeRuneInString(pattern[i:])
		if runeValue == '*' {
			regexp_pattern += `\Q`
			regexp_pattern += pattern[l:i]
			regexp_pattern += `\E`
			regexp_pattern += `.*`
			l = i + width
		} else if runeValue == '?' {
			regexp_pattern += `\Q`
			regexp_pattern += pattern[l:i]
			regexp_pattern += `\E`
			regexp_pattern += `.`
			l = i + width
		} else if i == len(pattern)-1 {
			regexp_pattern += `\Q`
			regexp_pattern += pattern[l:]
			regexp_pattern += `\E`
		}
		w = width
	}
	match, err := regexp.MatchString("^"+regexp_pattern+"$", target)
	return glisp.SexpBool(match), err
}

func stringContainsFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var str, substr string

	switch t := args[0].(type) {
	case glisp.SexpStr:
		str = string(t)
	default:
		//return glisp.SexpNull, errors.New("wildcard parttern must be string")
		return glisp.SexpNull, nil
	}

	switch t := args[1].(type) {
	case glisp.SexpStr:
		substr = string(t)
	default:
		//return glisp.SexpNull, errors.New("arg1 must be string")
		return glisp.SexpNull, nil
	}

	return glisp.SexpBool(strings.Contains(str, substr)), nil

}

func defStringFunctions(env *glisp.Glisp) {
	env.AddFunction("str-wcmatch?", stringWildcardMatchFunction)
	env.AddFunction("str-contains?", stringContainsFunction)
	shortcuts := `
         (defn str= [v1 v2] (and (string? v1) (string? v2) (= v1 v2)))
         (defn str!= [v1 v2] (not (str= v1 v2)))
         (defn str-empty? [s] (or (null? s) (and (string? s) (= s ""))))
         (defn str-not-empty? [s] (not (str-empty? s)))
         (defn str-not-contains? [str substr] (not (str-contains? str substr)))
         (defn str-not-wcmatch? [p s] (not (str-wcmatch? p s)))
    `
	env.EvalString(shortcuts)
}

func NewGlisp() *glisp.Glisp {
	env := glisp.NewGlisp()
	defmacroCondValues(env)
	defVersionCompareFunctions(env)
	defStringFunctions(env)
	return env
}
