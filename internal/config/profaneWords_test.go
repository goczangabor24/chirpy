package config

import (
	"fmt"
	"testing"
)

func TestReplaceProfaneWords(t *testing.T) {
	result1 := replaceProfaneWords("This Kerfuffle is sharbert finding itself in fornax.")

	fmt.Println(result1)

	if result1 != "This **** is **** finding itself in fornax." {
		t.Errorf("1st test failed")
	}

	result2 := replaceProfaneWords("This Kerfuffle is sharbert finding itself in fornax")

	fmt.Println(result2)

	if result2 != "This **** is **** finding itself in ****" {
		t.Errorf("failed")
	}

	result3 := replaceProfaneWords("This Kerfuffle is SHARbert finding itself in fORNAX")

	fmt.Println(result3)

	if result3 != "This **** is **** finding itself in ****" {
		t.Errorf("failed")
	}
}
