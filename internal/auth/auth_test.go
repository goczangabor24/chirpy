package auth

import (
	"fmt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	result, err := HashPassword("Szevasztok3")
	if err != nil {
		t.Errorf("failed")
	}
	fmt.Println(result)
}

func TestCheckPassword(t *testing.T) {
	hashedPassword, _ := HashPassword("Something else")

	result, err := CheckPassword("Szevasztok3", hashedPassword)
	if err != nil {
		t.Errorf("failed")
	}
	fmt.Println(result)
}
