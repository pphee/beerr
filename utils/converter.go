package utils

func BinaryConverter(number int, bits int) []int {
	result := make([]int, bits)
	for number > 0 && bits > 0 {
		bits--
		result[bits] = number % 2
		number /= 2
	}
	return result
}
