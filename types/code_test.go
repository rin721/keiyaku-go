package types

import "testing"

func TestCodeCategory(t *testing.T) {
	tests := map[Code]CodeCategory{
		CodeOK:                CategorySuccess,
		CodeInvalidArgument:   CategoryClient,
		CodeInvalidCredential: CategoryBusiness,
		CodeInternal:          CategorySystem,
	}
	for code, want := range tests {
		if got := code.Category(); got != want {
			t.Fatalf("Category(%d) = %s, want %s", code, got, want)
		}
	}
}

func TestNewResponseUsesDefaultMessage(t *testing.T) {
	got := NewResponse(CodeNotFound, "", nil)
	if got.Msg != MessageNotFound {
		t.Fatalf("Msg = %q, want %q", got.Msg, MessageNotFound)
	}
}
