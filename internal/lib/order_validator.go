package lib

import "strconv"

func CheckLuhn(ccn string) bool {
	sum := 0
	parity := len(ccn) % 2

	for i := 0; i < len(ccn); i++ {
		digit, err := strconv.Atoi(string(ccn[i]))
		if err != nil {
			return false
		}

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}
