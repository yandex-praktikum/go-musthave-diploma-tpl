package cookie

import (
	"testing"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/stretchr/testify/assert"
)

func TestCookieHandler_SetAndGet(t *testing.T) {
	handler := NewCookieHandler()

	user := &dto.UserData{
		Login: "test",
	}

	handler.Set("session1", user)

	got, ok := handler.Get("session1")

	assert.True(t, ok)
	assert.NotNil(t, got)
	assert.Equal(t, user.Login, got.Login)
}

func TestCookieHandler_Get_NotFound(t *testing.T) {
	handler := NewCookieHandler()

	got, ok := handler.Get("unknown")

	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestCookieHandler_Set_Overwrite(t *testing.T) {
	handler := NewCookieHandler()

	user1 := &dto.UserData{Login: "user1"}
	user2 := &dto.UserData{Login: "user2"}

	handler.Set("session1", user1)
	handler.Set("session1", user2)

	got, ok := handler.Get("session1")

	assert.True(t, ok)
	assert.Equal(t, "user2", got.Login)
}
