package luna

// CalculateLuhn return the check number
func CalculateLuhn(number int) int {
	checkNumber := checksum(number)

	if checkNumber == 0 {
		return 0
	}
	return 10 - checkNumber
}

func checksum(number int) int {
	var luna int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luna += cur
		number = number / 10
	}
	return luna % 10
}

// Luna check number is valid or not based on Luhn algorithm
func Luna(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}
