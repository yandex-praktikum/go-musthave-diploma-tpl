package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/luhn"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

func (h *Handler) GetBalance(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	curentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	balance, err := h.storage.GetBalance(curentuserID)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, balance)
}

func (h *Handler) Withdraw(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	curentuserID, err := getUserID(c)
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

	err = h.storage.Withdraw(curentuserID, withdraw)

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, "bonuses was debeted")
}

func (h *Handler) GetWithdraws(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	curentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	withdraws, err := h.storage.GetWithdraws(curentuserID)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if len(withdraws) == 0 {
		newErrorResponse(c, errors.New("NoContent"))
		return
	}
	c.JSON(http.StatusOK, withdraws)
}

type BalanceService struct {
	// log  logger.Logger
	repo repository.Balance
}

func NewBalanceStorage(repo repository.Balance) *BalanceService {
	return &BalanceService{repo: repo}
}

func (b *BalanceService) GetWithdraws(userID int) ([]models.WithdrawResponse, error) {
	return b.repo.GetWithdraws(userID)
}

func (b *BalanceService) GetBalance(userID int) (models.Balance, error) {
	return b.repo.GetBalance(userID)

}
func (b *BalanceService) Withdraw(userID int, withdraw models.Withdraw) error {

	numOrderInt, err := strconv.Atoi(withdraw.Order)
	if err != nil {
		return errors.New("PreconditionFailed")
	}

	correctnum := luhn.Valid(numOrderInt)

	if !correctnum {
		return errors.New("UnprocessableEntity")
	}

	balance, err := b.repo.GetBalance(userID)

	if err != nil {
		return err
	}
	// if balance.Current > withdraw.Sum {
	err := b.repo.DoWithdraw(userID, withdraw)

	if err != nil {
		return err
	}
	// } else {
	// 	return errors.New("PaymentRequired")
	// }

	return nil
}
