package wallet

import (
	"github.com/Muhamadi02/wallet/pkg/types"
	"github.com/google/uuid"
	"testing"
)

func TestService_FindAccountByID_success(t *testing.T) {
	svc := &Service{}

	account1, _ := svc.RegisterAccount("+992000000001")
	svc.RegisterAccount("+992000000002")
	svc.RegisterAccount("+992000000003")

	_, err := svc.FindAccountByID(account1.ID)
	if err != nil {
		t.Error(err)
	}

}

func TestService_FindAccountByID_fail(t *testing.T) {
	svc := &Service{}

	svc.RegisterAccount("+992000000001")
	svc.RegisterAccount("+992000000002")
	svc.RegisterAccount("+992000000003")

	_, err := svc.FindAccountByID(1000)
	if err == nil {
		t.Error(err)
	}
}

func TestService_FindPaymentById_success(t *testing.T) {
	svc := &Service{}

	account1, _ := svc.RegisterAccount("+992000000001")
	svc.RegisterAccount("+992000000002")

	err := svc.Deposit(account1.ID, 500_00)
	if err != nil {
		switch err {
		case ErrAmountMustBePositive:
			t.Error(err)
		case ErrAccountNotFound:
			t.Error(err)
		}
		return
	}

	payment, err1 := svc.Pay(account1.ID, 200_00, "auto")
	if err1 != nil {
		switch err1 {
		case ErrAmountMustBePositive:
			t.Error(err1)
		case ErrAccountNotFound:
			t.Error(err1)
		case ErrNotEnoughBalance:
			t.Error(err1)
		}
		return
	}

	_, err = svc.FindPaymentById(payment.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_FindPaymentById_fail(t *testing.T) {
	svc := &Service{}

	_, err := svc.FindPaymentById(uuid.New().String())
	if err != ErrPaymentNotFound {
		t.Error(err)
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := &Service{}

	account1, _ := svc.RegisterAccount("+992000000001")

	err := svc.Deposit(account1.ID, 500_00)
	if err != nil {
		t.Error(err)
		return
	}

	payment, err := svc.Pay(account1.ID, 200_00, "auto")
	if err != nil {
		t.Error(err)
	}

	err = svc.Reject(payment.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestService_Reject_fail(t *testing.T) {
	svc := &Service{}

	acc1, _ := svc.RegisterAccount("+992900000001")
	acc2, _ := svc.RegisterAccount("+992900000002")
	acc3, _ := svc.RegisterAccount("+992900000003")

	_ = svc.Deposit(acc1.ID, types.Money(100))
	_ = svc.Deposit(acc2.ID, types.Money(100))
	_ = svc.Deposit(acc3.ID, types.Money(100))

	svc.Pay(acc1.ID, types.Money(10), types.PaymentCategory("mobile"))
	svc.Pay(acc2.ID, types.Money(10), types.PaymentCategory("mobile"))
	svc.Pay(acc3.ID, types.Money(10), types.PaymentCategory("mobile"))

	err := svc.Reject(uuid.New().String())
	if err != ErrPaymentNotFound {
		t.Error(err)
	}
}
