package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type userEmail struct {
	Email string `json:"Email"`
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users []userEmail

func getUsers(r io.Reader) (users, error) {
	result := make(users, 0, 100_000)
	scanner := bufio.NewScanner(r)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var user userEmail
		if err := user.UnmarshalJSON(line); err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	domainPattern, err := regexp.Compile(`\.` + domain + `$`)
	if err != nil {
		return nil, err
	}

	for _, user := range u {
		atIdx := strings.IndexByte(user.Email, '@')
		if atIdx == -1 || atIdx == len(user.Email)-1 {
			continue
		}

		emailDomain := user.Email[atIdx+1:]

		if domainPattern.MatchString(emailDomain) {
			lowerDomain := strings.ToLower(emailDomain)
			result[lowerDomain]++
		}
	}
	return result, nil
}
