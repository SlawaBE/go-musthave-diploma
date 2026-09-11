package validator

import "strings"

func ValidateLuhn(number string) bool {
	number = strings.TrimSpace(number)
	if number == "" {
		return false
	}

	for _, ch := range number {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	digits := make([]int, len(number))
	for i, ch := range number {
		digits[i] = int(ch - '0')
	}

	parity := len(digits) % 2
	sum := 0

	for i, digit := range digits {
		if i%2 == parity {
			doubled := digit * 2
			if doubled > 9 {
				doubled -= 9
			}
			sum += doubled
		} else {
			sum += digit
		}
	}

	return sum%10 == 0
}
