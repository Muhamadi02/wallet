package wallet

import (
	"errors"
	"github.com/Muhamadi02/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be a greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64 // для генерации уникального номера аккаунта
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount

	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}
	return nil, ErrAccountNotFound
}

// FindPaymentByID возврашает платеж по идентификатору.
func (s *Service) FindPaymentById(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

// Reject возвращает платеж в случае ошибки.
func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentById(paymentID)
	if err != nil {
		return err
	}

	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount

	return nil
}

// Repeat повторяет платеж по идетификатору 
func (s *Service) Repeat(paymentID string)(*types.Payment, error){
	payment, err := s.FindPaymentById(paymentID)
	if err != nil {
		return nil, err
	}

	repeatPay, err := s.Pay(payment.AccountID,payment.Amount, payment.Category)
	if err != nil{
		return nil, err
	}

	return repeatPay, nil
}

// FavoritePayment создает избранное из конкретного платежа
func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentById(paymentID)
	if err != nil {
		return nil, err
	}

	favPaymentID := uuid.New().String()
	favPayment := &types.Favorite{
		ID: favPaymentID,
		AccountID: payment.AccountID,
		Name: name,
		Amount: payment.Amount,
		Categoty: payment.Category,
	}

	s.favorites = append(s.favorites, favPayment)
	return favPayment, nil
}

// FindFavoriteByID - поиск избранного платежа по идентификатору.
func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

// PayFromFavorite - совершает платеж из конкретного избранного
func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favPayment, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(favPayment.AccountID, favPayment.Amount, favPayment.Categoty)
	if err != nil {
		return nil, err
	}

	return payment, nil
}