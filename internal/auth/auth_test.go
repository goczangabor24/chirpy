package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	result, err := HashPassword("Something")
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

func TestMakeJWT(t *testing.T) {
	duration := 2 * time.Second
	fmt.Println(MakeJWT(uuid.New(), "mySecret", time.Duration(duration)))
}
