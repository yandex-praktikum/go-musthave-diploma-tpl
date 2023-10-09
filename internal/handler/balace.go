package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

func (h *Handler) GetBalance(c *gin.Context) {
	curentuserId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	balance, err := h.storage.GetBalance(curentuserId)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, balance)
}

func (h *Handler) Withdraw(c *gin.Context) {
	curentuserId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	var withdraw models.Withdraw

	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		newErrorResponse(c, err)
		h.log.Error(err)
		return
	}
	defer c.Request.Body.Close()

	if err := json.Unmarshal(jsonData, &withdraw); err != nil {
		newErrorResponse(c, err)
		h.log.Error(err)
		return
	}

	err = h.storage.Withdraw(curentuserId, withdraw)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, "bonuses was debeted")
}

type BalanceService struct {
	// log  logger.Logger
	repo repository.Balance
}

func NewBalanceStorage(repo repository.Balance) *BalanceService {
	return &BalanceService{repo: repo}
}

func (b *BalanceService) GetBalance(user_id int) (models.Balance, error) {
	return b.repo.GetBalance(user_id)

}
func (b *BalanceService) Withdraw(user_id int, withdraw models.Withdraw) error {
	balance, err := b.repo.GetBalance(user_id)

	if err != nil {
		return err
	}
	if balance.Current > withdraw.Sum {
		err := b.repo.DoWithdraw(user_id, withdraw)

		if err != nil {
			return err
		}
	} else {
		return errors.New("PaymentRequired")
	}

	return nil
}
