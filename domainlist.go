package main

import (
	"bufio"
	"github.com/pkg/errors"
	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"golang.org/x/text/encoding"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type DomainList struct {
	Name                string
	FullDomains         []*router.Domain
	RegexpDomains       []*router.Domain
	DomainDomains       []*router.Domain
	UniqueDomainDomains []*router.Domain
	GeoSite             *router.GeoSite
}

func (l *DomainList) push(rule *router.Domain) {
	switch rule.Type {
	case router.Domain_Full:
		l.FullDomains = append(l.FullDomains, rule)
	case router.Domain_RootDomain:
		l.DomainDomains = append(l.DomainDomains, rule)
	case router.Domain_Regex:
		l.RegexpDomains = append(l.RegexpDomains, rule)
	}
}

func (l *DomainList) parseDomainType(domain string) router.Domain_Type {
	if strings.Contains(domain, "*") {
		return router.Domain_Regex
	}
	scheme := regexp.MustCompile("^[a-zA-Z]+://")
	if scheme.MatchString(domain) {
		return router.Domain_Full
	}
	return router.Domain_RootDomain
}

func (l *DomainList) wildcardToRegex(domain string) (string, error) {
	_, parts := splitAndTrim(strings.ReplaceAll(strings.ReplaceAll(domain, "https://", ""), "http://", ""), "/")
	_, parts = splitAfterAndCount(parts[0], "*.")
	var sb strings.Builder
	for _, column := range parts {
		if column == "*." {
			sb.WriteString("(.*\\.)?")
			continue
		}
		sb.WriteString(regexp.QuoteMeta(column))
	}
	return sb.String(), nil
}

func (l *DomainList) parseDomain(domain string) (*router.Domain, error) {
	if isBlank(domain) {
		return nil, nil
	}
	domainType := l.parseDomainType(domain)
	var sb strings.Builder
	switch domainType {
	case router.Domain_Regex:
		parsed, err := l.wildcardToRegex(domain)
		if err != nil {
			return nil, err
		}
		if isBlank(parsed) {
			return nil, nil
		}
		sb.WriteString(parsed)
	case router.Domain_Full:
		sb.WriteString(domain)
	case router.Domain_RootDomain:
		sb.WriteString(domain)
	default:
		return nil, errors.New("unexpected domain type: " + domainType.String())
	}
	domain = sb.String()
	var routerDomain router.Domain
	routerDomain.Type = domainType
	routerDomain.Value = domain
	return &routerDomain, nil
}

func (l *DomainList) parseDomains(line string) ([]*router.Domain, error) {
	if isBlank(line) {
		return nil, nil
	}
	_, parts := splitAndTrim(line, "|")
	var domains []*router.Domain
	for _, domain := range parts {
		routerDomain, err := l.parseDomain(domain)
		if err != nil {
			return nil, err
		}
		if routerDomain == nil {
			continue
		}
		domains = append(domains, routerDomain)
	}
	return domains, nil
}

func (l *DomainList) parseRule(line string) ([]*router.Domain, error) {
	if isBlank(line) {
		return nil, errors.New("line is empty")
	}
	count, columns := splitAndTrim(line, ";")
	if count <= 1 {
		return nil, errors.New("line is missing delimiters")
	}
	var domains []*router.Domain
	for index, column := range columns[:count-2] {
		if index == 0 {
			continue
		}
		parsed, err := l.parseDomains(column)
		if err != nil {
			return nil, err
		}
		domains = append(domains, parsed...)
	}
	return domains, nil
}

func (l *DomainList) Flatten() error {
	sort.Slice(l.DomainDomains, func(i, j int) bool {
		return len(strings.Split(l.DomainDomains[i].GetValue(), ".")) < len(strings.Split(l.DomainDomains[j].GetValue(), "."))
	})
	trie := NewDomainTrie()
	for _, domain := range l.DomainDomains {
		success, err := trie.Insert(domain.GetValue())
		if err != nil {
			return err
		}
		if success {
			l.UniqueDomainDomains = append(l.UniqueDomainDomains, domain)
		}
	}
	return nil
}

func (l *DomainList) ToGeoSites() *router.GeoSiteList {
	domainList := new(router.GeoSiteList)
	domain := new(router.GeoSite)
	domain.CountryCode = l.Name
	domain.Domain = append(domain.Domain, l.FullDomains...)
	domain.Domain = append(domain.Domain, l.UniqueDomainDomains...)
	domain.Domain = append(domain.Domain, l.RegexpDomains...)
	domain.Domain = append(domain.Domain)
	l.GeoSite = domain
	domainList.Entry = append(domainList.Entry, l.GeoSite)
	return domainList
}

func (l *DomainList) ToPlainText() []byte {
	bytes := make([]byte, 0, 1024*512)
	for _, rule := range l.GeoSite.Domain {
		ruleVal := strings.TrimSpace(rule.GetValue())
		if len(ruleVal) == 0 {
			continue
		}
		var ruleString string
		switch rule.Type {
		case router.Domain_Full:
			ruleString = "full:" + ruleVal
		case router.Domain_RootDomain:
			ruleString = "domain:" + ruleVal
		case router.Domain_Regex:
			ruleString = "regexp:" + ruleVal
		}
		bytes = append(bytes, []byte(ruleString+"\n")...)
	}
	return bytes
}

func readAndParseLine(decoder *encoding.Decoder, scanner *bufio.Scanner, list *DomainList, lineNum int) (int, error) {
	if err := scanner.Err(); err != nil {
		return lineNum, err
	}
	if !scanner.Scan() {
		return lineNum, nil
	}
	utf8, err := decoder.Bytes(scanner.Bytes())
	if err != nil {
		return lineNum, err
	}
	line := strings.TrimSpace(string(utf8))
	if isEmpty(line) {
		return readAndParseLine(decoder, scanner, list, lineNum+1)
	}
	parsedRules, err := list.parseRule(line)
	if err != nil {
		return lineNum, err
	}
	for _, rule := range parsedRules {
		list.push(rule)
	}
	return readAndParseLine(decoder, scanner, list, lineNum+1)
}

func NewDomainList(name string) *DomainList {
	return &DomainList{
		Name:                name,
		FullDomains:         make([]*router.Domain, 0, 10),
		RegexpDomains:       make([]*router.Domain, 0, 10),
		DomainDomains:       make([]*router.Domain, 0, 10),
		UniqueDomainDomains: make([]*router.Domain, 0, 10),
	}
}

func (l *DomainList) unmarshalCSV(decoder *encoding.Decoder, file *os.File) (int, error) {
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	if scanner.Scan() && scanner.Err() == nil {
		if lineNum, err := readAndParseLine(decoder, scanner, l, 2); err != nil {
			return lineNum, err
		}
	}
	return 1, scanner.Err()
}

func Unmarshal(decoder *encoding.Decoder, listName string, path string) (*DomainList, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}(file)
	list := NewDomainList(listName)
	if lineNum, err := list.unmarshalCSV(decoder, file); err != nil {
		ex := errors.Wrap(err, "could not parse rule at line number: "+strconv.Itoa(lineNum))
		return nil, ex
	}
	return list, nil
}
