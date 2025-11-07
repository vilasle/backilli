package model

import (
	"net/url"
	"regexp"
)

type Source interface {
}

type FileSource struct {
	path         string
	filterRegexp *regexp.Regexp
}

type PostgresSource struct {
	name string
	uri  *url.URL
}
