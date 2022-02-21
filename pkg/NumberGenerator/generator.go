package numbergenerator

import (
	crypto "crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

func GenerateNumber(len int) (string, error) {
	var sum string
	arr := make([]int, len)
	for i := range arr {
		arr[i] = 9
		sum = fmt.Sprint(sum, arr[i])
	}
	res, _ := strconv.ParseInt(sum, 10, 64)
	number, err := crypto.Int(crypto.Reader, big.NewInt(int64(res)))
	if err != nil {
		return "", err
	}
	return number.String(), nil
}
