package gfw

import (
	"fmt"
	"io/ioutil"
	"os"

	"net/http"

	"strings"
)

type TrieNode struct {
	children      map[string]*TrieNode
	isEndOfDomain bool
}

type CheckDomain struct {
	root *TrieNode
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[string]*TrieNode),
	}
}

func newCheckDomain() *CheckDomain {
	return &CheckDomain{
		root: NewTrieNode(),
	}
}
func (cd *CheckDomain) Inserts(domains []string) {
	for _, domain := range domains {
		cd.Insert(domain)
	}
}
func (cd *CheckDomain) Insert(domain string) {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	node := cd.root
	for _, part := range parts {
		if _, ok := node.children[part]; !ok {
			node.children[part] = NewTrieNode()
		}
		node = node.children[part]
	}
	node.isEndOfDomain = true
}

func (cd *CheckDomain) IsSubdomainOfAny(domain string) bool {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	node := cd.root
	for _, part := range parts {
		if child, ok := node.children[part]; ok {
			node = child
			if node.isEndOfDomain {
				return true
			}
		} else {
			return false
		}
	}
	return false
}

func NewDomains(url string) *CheckDomain {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	gfwList := strings.Split(string(body), "\n")
	newObj := newCheckDomain()
	fmt.Println("len: ", len(gfwList))
	newObj.Inserts(gfwList)
	return newObj
}

func NewDomainsFromFile(url string) *CheckDomain {
	resp, err := os.Open(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Close()

	body, err := ioutil.ReadAll(resp)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	gfwList := strings.Split(string(body), "\n")
	newObj := newCheckDomain()
	fmt.Println("len: ", len(gfwList))
	newObj.Inserts(gfwList)
	return newObj
}
