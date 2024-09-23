package cli

import (
	"os"
	"strings"
	"testing"
)

func TestGetUserInputString(t *testing.T) {
	// モック標準入力データを準備
	input := "test_user_input\n"
	r := strings.NewReader(input)

	// os.Stdin を一時的に置き換える
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
}
