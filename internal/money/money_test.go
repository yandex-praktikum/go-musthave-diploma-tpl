package money

import (
	"encoding/json"
	"testing"
)

func TestParseAndFormat(t *testing.T) {
	tests := map[string]string{
		"0":      "0",
		"42":     "42",
		"500.50": "500.5",
		"10.05":  "10.05",
	}

	for raw, want := range tests {
		amount, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", raw, err)
		}
		if got := amount.String(); got != want {
			t.Fatalf("Amount.String() = %q, want %q", got, want)
		}
	}
}

func TestAmountJSON(t *testing.T) {
	var amount Amount
	if err := json.Unmarshal([]byte(`500.5`), &amount); err != nil {
		t.Fatal(err)
	}
	if amount != 50050 {
		t.Fatalf("amount = %d, want 50050", amount)
	}

	data, err := json.Marshal(amount)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "500.5" {
		t.Fatalf("json = %s, want 500.5", data)
	}
}
