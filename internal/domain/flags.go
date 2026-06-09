package domain

import "fmt"

type UrlsFlag []string

func (uf *UrlsFlag) String() string {
	return fmt.Sprintf("%v", *uf)
}

func (uf *UrlsFlag) Set(value string) error {
	*uf = append(*uf, value)
	return nil
}

type BranchesFlag []string

func (bf *BranchesFlag) String() string {
	return fmt.Sprintf("%v", *bf)
}

func (bf *BranchesFlag) Set(value string) error {
	*bf = append(*bf, value)
	return nil
}
