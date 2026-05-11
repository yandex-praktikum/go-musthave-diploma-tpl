package models_test

import (
	"encoding/json"
	"testing"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/stretchr/testify/require"
)

// TestFromJSON_Invalid проверяет обработку невалидного JSON
func TestFromJSON_Invalid(t *testing.T) {
	invalidJSON := `{"id": "123", "login": ` // Оборванный JSON

	_, err := models.FromJSON[models.User]([]byte(invalidJSON))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON, but got nil")
	}
}

func TestAccrual_MarshalUnmarshal_Cycle(t *testing.T) {
	original := models.Accrual(17348) // 173.48

	// Marshal (должно превратиться в 173.48)
	data, err := json.Marshal(original)
	require.NoError(t, err)
	require.Equal(t, "173.48", string(data))

	// Unmarshal (должно превратиться обратно в 17348)
	var restored models.Accrual
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	require.Equal(t, original, restored)
}

func TestAccrual_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    models.Accrual
		wantErr bool
	}{
		{
			name:  "integer input",
			input: `500`,
			want:  50000, // 500.00 -> 50000 копеек
		},
		{
			name:  "float input exact",
			input: `5.00`,
			want:  500, // 5.00 -> 500 копеек
		},
		{
			name:  "float input fractional",
			input: `12.34`,
			want:  1234, // 12.34 -> 1234 копейки
		},
		{
			name:  "float input with rounding up",
			input: `12.345`, // 12.345 * 100 = 1234.5 -> округляем до 1235
			want:  1235,
		},
		{
			name:  "float input with rounding down",
			input: `12.344`, // 12.344 * 100 = 1234.4 -> округляем до 1234
			want:  1234,
		},
		{
			name:  "negative number",
			input: `-10.50`,
			want:  -1050,
		},
		{
			name:  "null input",
			input: `null`,
			want:  0, // null должен обрабатываться как 0
		},
		{
			name:    "invalid input string",
			input:   `"abc"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc models.Accrual
			err := json.Unmarshal([]byte(tt.input), &acc)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, acc, "accrual value mismatch")
			}
		})
	}
}
