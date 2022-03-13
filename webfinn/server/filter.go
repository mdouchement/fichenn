package server

import (
	"regexp"
)

type Rule struct {
	Method  string `json:"method"`
	Route   string `json:"route"`
	matcher *regexp.Regexp
}

type Filter struct {
	rules []*Rule
}

func NewFilter(rules []*Rule) (*Filter, error) {
	var err error
	for _, r := range rules {
		r.matcher, err = regexp.Compile(r.Route)
		if err != nil {
			return nil, err
		}
	}

	return &Filter{
		rules: rules,
	}, nil
}

func (f *Filter) Match(method, route string) bool {
	for _, r := range f.rules {
		if r.Method == method && r.matcher.MatchString(route) {
			return true
		}
	}
	return false
}
