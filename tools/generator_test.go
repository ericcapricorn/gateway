package tools

import (
	"fmt"
	"testing"
)

func TestKeyGenerate(t *testing.T) {
	var generator RSAGenarator
	key, err := generator.Generate(256)
	if err != nil {
		t.Error("generator key failed")
	}
	fmt.Println("PublicKey:", key.PublicKey)
}
