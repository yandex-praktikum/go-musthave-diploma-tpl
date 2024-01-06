package general

import (
	"errors"
	"fmt"
	"syscall"
	"time"
)

func RetryCode(f func() error, errorType syscall.Errno) error {
	sleepTime := time.Second
	countOfRepetition := 3
	for i := 0; i >= 0; i++ {
		err := f()
		if err != nil {
			fmt.Println("Error:", err.Error())
			isErrorType := errors.Is(err, errorType)
			fmt.Println("isErrorType:", isErrorType)
			if isErrorType && i < countOfRepetition {
				time.Sleep(sleepTime)
				fmt.Println("Repeating... SleepTime:", sleepTime)
				sleepTime += 2 * time.Second
				continue
			}
			return err
		}
		return nil
	}
	return nil
}
