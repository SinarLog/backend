package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// I code this by myself

var (
	unique = []uint16{
		35,
		36,
		38,
		42,
		63,
		64,
	}

	partitionLength = [...]uint16{
		5,
		6,
		7,
	}

	/*
		Unique Characters are [35, 36, 38, 42, 63, 64]
		Connectors of partitions are 45
		Lowercase letters are 97-122
		Uppercase leters are 65-90
	*/

	/*
		According to the ASCII codes, the variables:
		- maxLower and minLower are used to get random lowercase letters,
		- maxUpper and minUpper are used to get random uppercase letters,
		- maxNumber and minNumber are used to get random numbers.

		Meanwhile, picker is a variable to choose which of lowercase, uppercase or
		unique characters should it add next.
	*/
	maxLower, minLower, maxUpper, minUpper, maxNumber, minNumber, picker int = 122, 97, 90, 65, 57, 48, 4
)

func GenerateRandomPassword() string {
	/*
		- Each password must have 3 partitions.
		- Each password must contain 1 unique character, 1 uppercase letter, and 1 number.
		- Each partitions must be of length 5-7 characters.
	*/

	// Create variable for the while loop and result.
	var result string

	// Create and seed the generator.
	// Typically a non-fixed seed should be used, such as time.Now().In(utils.CURRENT_LOC).UnixNano().
	// Using a fixed seed will produce the same output on every run.
	r := rand.New(rand.NewSource(50))
	l := rand.New(rand.NewSource(time.Now().In(CURRENT_LOC).UnixNano()))
	eachPartitionLength := r.Perm(len(partitionLength))

	for index, value := range eachPartitionLength {
		// Create variable for the for loop of type uint16.
		var i uint16
		for i = 0; i < partitionLength[value]; i++ {
			switch r.Intn(picker) {
			case 2:
				result += string(rune(unique[l.Intn(len(unique))]))
			case 3:
				result += string(rune(l.Intn(maxLower-minLower+1) + minLower))
			case 1:
				result += string(rune(l.Intn(maxUpper-minUpper+1) + minUpper))
			case 0:
				result += string(rune(l.Intn(maxNumber-minNumber+1) + minNumber))
			}
		}
		if index != len(eachPartitionLength)-1 {
			result += string(rune(45))
		}
		continue
	}

	return result
}

func ToPatternMatching(str string) string {
	return "%" + str + "%"
}

func ToOrderSQL(str1, str2 string) string {
	return str1 + " " + str2
}

func GetFirstNameFromFullName(fullName string) string {
	return strings.Split(fullName, " ")[0]
}

func PadIntegerTime(n int) string {
	if n > 9 {
		return strconv.Itoa(n)
	} else {
		return "0" + strconv.Itoa(n)
	}
}

func IsStrongPassword(s string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 7 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}
