package wallet

import (
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
