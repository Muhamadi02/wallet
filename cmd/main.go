package main

import (
	"fmt"
	"github.com/Muhamadi02/wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = svc.Deposit(account.ID, 10)
	if err != nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")
		}
		return
	}

	fmt.Println(account.Balance)

}
