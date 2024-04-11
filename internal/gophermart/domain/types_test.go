package domain_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckLuhn(t *testing.T) {
	number := "5062821234567892"
	require.True(t, domain.CheckLuhn(number))

	number = "5062821234567893"
	require.False(t, domain.CheckLuhn(number))
}

func TestRFC3339TimeSerialization(t *testing.T) {

	var tm time.Time
	r := domain.TimePtr(tm)
	b, err := r.MarshalJSON()
	require.NotNil(t, b)

	val := strings.Trim(string(b), `"`)
	require.Equal(t, "0001-01-01T00:00:00Z", val)

	require.NoError(t, err)

	val = "2020-12-09T16:09:57+03:00"
	err = r.UnmarshalJSON([]byte(val))
	require.NoError(t, err)

	tm, err = time.Parse(time.RFC3339, val)
	require.NoError(t, err)

	require.True(t, tm.Equal(time.Time(*r)))
}

func TestJsonRFC3339Serialization(t *testing.T) {

	oData := domain.WithdrawData{
		Sum:   500,
		Order: "2377225624",
	}

	res, err := json.Marshal(oData)
	require.NoError(t, err)

	require.JSONEq(t, `{"order": "2377225624",
          "sum": 500}`, string(res))

	var oData2 domain.WithdrawData
	err = json.Unmarshal(res, &oData2)
	require.NoError(t, err)

	assert.True(t, reflect.DeepEqual(oData, oData2))
}

func TestAccrualDataDeserialization(t *testing.T) {

	var acrualData domain.AccrualData

	err := json.Unmarshal([]byte(`{"number": "2377225624","status": "PROCESSED","accrual": 500}`), &acrualData)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(domain.AccrualData{
		Number:  "2377225624",
		Status:  "PROCESSED",
		Accrual: domain.Float64Ptr(500),
	}, acrualData))

	err = json.Unmarshal([]byte(`{"number": "2377225624","status": "PROCESSEDDDD","accrual": 500}`), &acrualData)
	require.ErrorIs(t, err, domain.ErrDataFormat)
}
