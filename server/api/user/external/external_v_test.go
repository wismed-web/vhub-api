package external

import (
	"fmt"
	"testing"
)

func TestVUserExists(t *testing.T) {
	ok, err := ExtUserExistsCheck("13980824611")
	fmt.Println(ok, err)
}

func TestVUserLogin(t *testing.T) {
	ok, err := ExtUserLoginCheck("13980824611", "123456")
	fmt.Println(ok, err)
	ok, err = ExtUserLoginCheck("13980824611", "12345")
	fmt.Println(ok, err)
}
