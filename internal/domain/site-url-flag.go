package domain

import "fmt"

type UrlsFlag []string

func (suf *UrlsFlag) String() string {
	return fmt.Sprintf("%v", *suf)
}

func (suf *UrlsFlag) Set(value string) error {
	*suf = append(*suf, value)
	return nil
}
