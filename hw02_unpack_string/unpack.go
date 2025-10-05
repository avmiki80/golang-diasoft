package hw02unpackstring

import (
	"errors"
	"strconv"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func isValidEscapeTarget(r rune) bool {
	return r == '\\' || unicode.IsDigit(r)
}

func handleEscapeSequence(runes []rune, index int) (rune, int, error) {
	if index+1 >= len(runes) {
		return 0, index, ErrInvalidString
	}

	nextRune := runes[index+1]
	if !isValidEscapeTarget(nextRune) {
		return 0, index, ErrInvalidString
	}

	return nextRune, index + 1, nil
}

func parseDigits(runes []rune, startIndex int) (int, int, error) {
	j := startIndex

	for j < len(runes) && unicode.IsDigit(runes[j]) {
		j++
	}

	result, err := strconv.Atoi(string(runes[startIndex:j]))
	if err != nil {
		return 0, startIndex, ErrInvalidString
	}

	return result, j - 1, nil
}

func applyRepetition(result []rune, count int) []rune {
	if count == 0 {
		if len(result) > 0 {
			return result[:len(result)-1]
		}
		return result
	}

	if count == 1 {
		return result
	}

	prevRune := result[len(result)-1]
	currentLen := len(result)
	result = append(result, make([]rune, count-1)...)
	for i := currentLen; i < len(result); i++ {
		result[i] = prevRune
	}

	return result
}

func Unpack(s string) (string, error) {
	var result []rune
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		currentRune := runes[i]
		switch {
		case currentRune == '\\':
			nextRune, newIndex, err := handleEscapeSequence(runes, i)
			if err != nil {
				return "", err
			}
			result = append(result, nextRune)
			i = newIndex
		case unicode.IsDigit(currentRune):
			if i == 0 {
				return "", ErrInvalidString
			}
			countRepetition, newIndex, err := parseDigits(runes, i)
			if err != nil || countRepetition >= 10 {
				return "", ErrInvalidString
			}
			result = applyRepetition(result, countRepetition)
			i = newIndex
		default:
			result = append(result, currentRune)
		}
	}
	return string(result), nil
}
