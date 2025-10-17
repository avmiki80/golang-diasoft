package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type Word struct {
	Text  string
	Count int
}

type Words []Word

var taskWithAsteriskIsCompleted = true

func Top10(s string) []string {
	if s == "" {
		return nil
	}

	strArr := PrepareData(s)
	if strArr == nil {
		return nil
	}

	wordsMap := CreateWordMap(strArr, taskWithAsteriskIsCompleted)

	if len(wordsMap) == 0 {
		return nil
	}

	return NewWords(wordsMap).Sort().BuildResult()
}

func PrepareData(s string) []string {
	return strings.Fields(s)
}

func CreateWordMap(strArr []string, taskWithAsteriskIsCompleted bool) map[string]int {
	wordMap := make(map[string]int)
	for _, s := range strArr {
		if taskWithAsteriskIsCompleted {
			s = PrepareWord(s)
			if s == "" || s == "-" {
				continue
			}
		}
		wordMap[s]++
	}
	return wordMap
}

func PrepareWord(word string) string {
	cleaned := strings.TrimSpace(strings.ToLower(word))
	cleaned = strings.Trim(cleaned, "!@#$%^&*()_+=[]{}|;:,.<>?/\\ \"'`~")
	return cleaned
}

func NewWords(wordMap map[string]int) Words {
	words := make(Words, 0, len(wordMap))

	for k, v := range wordMap {
		words = append(words, Word{k, v})
	}
	return words
}

func (words Words) Sort() Words {
	if len(words) == 0 {
		return words
	}
	sort.SliceStable(words, func(i, j int) bool {
		if words[i].Count != words[j].Count {
			return words[i].Count > words[j].Count
		}
		return words[i].Text < words[j].Text
	})
	return words
}

func (words Words) BuildResult() []string {
	if len(words) == 0 {
		return nil
	}
	limit := len(words)
	if limit >= 10 {
		limit = 10
	}

	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = words[i].Text
	}

	return result
}
