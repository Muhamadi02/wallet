package wallet

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/Muhamadi02/wallet/pkg/types"
	"github.com/google/uuid"
)

type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, []*types.Favorite, error) {
	// регистрируем там пользователя
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// пополняем его счёт
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	//выполняем платежи
	//можем создать слайс нужной длины, поскольку знаем размер
	payments := make([]*types.Payment, len(data.payments))
	favorites := make([]*types.Favorite, len(data.payments))
	for i, payment := range data.payments {
		// тогда здесь работаем просто через index, а не через append
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}

		favPaymentName := "Favorite payment_" + strconv.Itoa(i)
		favorites[i], err = s.FavoritePayment(payments[i].ID, favPaymentName)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("can't make favorite payment, error = %v", err)
		}
	}

	return account, payments, favorites, nil
}

var defaultTestAccount = testAccount{
	phone:   "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

var defaultTestAccount2 = testAccount{
	phone:   "+992000000002",
	balance: 15_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 2_500_00, category: "auto"},
		{amount: 3_000_00, category: "phone"},
		{amount: 5_000_00, category: "shop"},
	},
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	//регистрируем там пользователя
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}

	//пополняем его счёт
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	return account, nil
}

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
	//создаём сервис
	s := newTestService()

	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// попробуем найти платёж
	payment := payments[0]
	got, err := s.FindPaymentById(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID():  error = %v", err)
		return
	}

	//сравниваем платежи
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return
	}
}

func TestService_FindPaymentById_fail(t *testing.T) {
	// создаём сервис
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// пробуем найти несуществующий платёж
	_, err = s.FindPaymentById(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentyID(): must return error, returned nil")
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	// создаём сервис
	s := newTestService()
	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//попробуем отменить платеж
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := s.FindPaymentById(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error = %v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changed, payment = %v", savedPayment)
		return
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}

	/*
		//регистрируем там пользователя
		account1, err := svc.RegisterAccount("+992000000001")
		if err != nil {
			t.Errorf("Reject(): can't register account, error = %v", err)
			return
		}

		// пополняем его счёт
		err = svc.Deposit(account1.ID, 10_000_00)
		if err != nil {
			t.Errorf("Reject(): can't deposit account, error = %v", err)
			return
		}

		//осуществляем платеж на его счёт
		payment, err := svc.Pay(account1.ID, 1000_00, "auto")
		if err != nil {
			t.Errorf("Reject(): can't create payment, error = %v", err)
		}

		//попробуем отменить платёж
		err = svc.Reject(payment.ID)
		if err != nil {
			t.Errorf("Reject(): error = %v", err)
			return
		}*/
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

func TestService_Repeat_success(t *testing.T) {
	s := newTestService()
	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// попробуем повторить платеж
	payment := payments[0]
	got, err := s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Repeat(): error : %v", err)
		return
	}

	if got.AccountID != payment.AccountID {
		t.Errorf("Repeat(): repeat account is not payment account, \n Repeated payment = %v,\n Rejected payment = %v", got, payment)
		return
	}
	if got.Amount != payment.Amount {
		t.Errorf("Repeat(): repeat amount don't equal payment amount, \n Repeated payment = %v,\n Rejected payment = %v", got, payment)
		return
	}
	if got.Category != payment.Category {
		t.Errorf("Repeat(): repeat category don't equal payment category, \n Repeated payment = %v,\n Rejected payment = %v", got, payment)
		return
	}
	if got.Status != payment.Status {
		t.Errorf("Repeat(): repeat status don't equal payment status,\n Repeated payment = %v,\n Rejected payment = %v", got, payment)
		return
	}

}

func TestService_Repeat_notFound(t *testing.T) {
	s := newTestService()

	_, _, _, err := s.addAccount(defaultTestAccount)

	if err != nil {
		t.Error(err)
		return
	}

	payment := uuid.New().String()
	_, err = s.Repeat(payment)
	if err == nil {
		t.Errorf("Repeat(): must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("Repeat(): must return ErrPaymentNotFound, returned: %v", err)
		return
	}

}

func TestService_FavoritePayment_success(t *testing.T) {
	s := newTestService()

	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, "megafon")
	if err != nil {
		t.Errorf("FavoritePayment(): error: %v", err)
		return
	}
}

func TestService_Favorite_notFound(t *testing.T) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	favoriteID := uuid.New().String()
	_, err = s.FavoritePayment(favoriteID, "my favorite payment")
	if err == nil {
		t.Errorf("FavoritePayment(): must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FavoritePayment(): must return ErrPaymentNotFound, returned: %v", err)
		return
	}
}

func TestService_FindFavoriteByID_success(t *testing.T) {
	s := newTestService()

	_, _, favorites, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error("FindFavoriteByID(): must return error, returned nil")
		return
	}

	favorite := favorites[0]
	result, err := s.FindFavoriteByID(favorite.ID)
	if err != nil {
		t.Errorf("FindFavoriteByID(): error: %v", err)
		return
	}

	if !reflect.DeepEqual(result, favorite) {
		t.Errorf("FindFavoriteByID(): wrong favorite payment returned = %v", err)
		return
	}
}

func TestService_FindFavoriteByID_notFound(t *testing.T) {
	s := newTestService()

	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	favoriteID := uuid.New().String()
	_, err = s.FindFavoriteByID(favoriteID)
	if err == nil {
		t.Error("FindFavoriteByID(): must return error, returned nil")
		return
	}
	if err != ErrFavoriteNotFound {
		t.Errorf("FindFavoriteByID(): must return ErrFavoriteNotFound, returned: %v", err)
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	s := newTestService()
	_, _, favorites, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	favorite := favorites[0]
	payment, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): error : %v", err)
		return
	}

	if payment.AccountID != favorite.AccountID {
		t.Errorf("PayFromFavorite(): account ID's difference, \n Current payment = %v, \n favorite payment = %v", payment, favorite)
		return
	}

	if payment.Amount != favorite.Amount {
		t.Errorf("PayFromFavorite(): amount of payment difference,\n Current payment = %v,\n favorite payment = %v", payment, favorite)
		return
	}

	if payment.Category != favorite.Categoty {
		t.Errorf("PayFromFavorite(): category of payment difference, \n Current payment = %v, \n favorite payment = %v", payment, favorite)
		return
	}
}

func TestService_PayFromFavorite_notFound(t *testing.T) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	favID := uuid.New().String()
	_, err = s.PayFromFavorite(favID)
	if err == nil {
		t.Error("PayFromFavorite(): must return error, returned nil")
		return
	}
	if err != ErrFavoriteNotFound {
		t.Errorf("PayFromFavorite(): must return ErrFavoriteNotFound, returned: %v", err)
		return
	}
}

func TestService_ExportToFile_success(t *testing.T) {
	s := newTestService()

	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	err = s.ExportToFile("data/accounts.txt")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_ExportToFile_notFound(t *testing.T) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	err = s.ExportToFile("")
	if err == nil {
		t.Error(err)
		return
	}
}

func TestService_ImportFromFile_success(t *testing.T) {
	s := newTestService()

	err := s.ImportFromFile("data/accounts.txt")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_ImportFromFile_notFound(t *testing.T) {
	s := newTestService()

	err := s.ImportFromFile("")
	if err == nil {
		t.Error(err)
		return
	}
}

func TestService_Export_full(t *testing.T) {
	s := newTestService()
	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, "my favor payment")
	if err != nil {
		t.Errorf("FavoritePayment(): error: %v", err)
		return
	}

	err = s.Export("data")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_Import_full_success(t *testing.T) {
	s := newTestService()

	err := s.Import("data")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_ExportAccountHistory_success(t *testing.T) {
	s := newTestService()

	acc, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = s.ExportAccountHistory(acc.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_ExportAccountHistory_notFound(t *testing.T) {
	s := newTestService()

	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	tempID := s.nextAccountID + 1
	_, err = s.FindAccountByID(tempID)
	if err == nil {
		t.Error("FindAccountByID(): must return error, returned nil")
		return
	}

	_, err = s.ExportAccountHistory(tempID)
	if err == nil {
		t.Error("ExportAccountHistory(): must return error, returned nil")
		return
	}
	if err != ErrAccountNotFound {
		t.Errorf("ExportAccountsHistory(): must return ErrAccountNotFound, returned %v", err)
		return
	}
}

func TestService_HistoryToFiles_success(t *testing.T) {
	s := newTestService()

	acc, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payments, err := s.ExportAccountHistory(acc.ID)
	if err != nil {
		t.Error(err)
	}
	err = s.HistoryToFiles(payments, "data", 3)
	if err != nil {
		t.Error(err)
	}
}

func TestService_HistoryToFiles_notSuccess(t *testing.T) {
	s := newTestService()

	payment := []types.Payment{}
	err := s.HistoryToFiles(payment, "data", 12)
	if err != nil {
		t.Error(err)
	}
}

func TestService_SumPayments(t *testing.T) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	sum := s.SumPayments(0)
	if sum != 1_000_00 {
		t.Errorf("SumPayments(): sum = %v", sum)
	}

}

func BenchmarkSumPayments(b *testing.B) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		b.Error(err)
	}
	want := types.Money(1_000_00)
	for i := 0; i < b.N; i++ {
		result := s.SumPayments(3)
		if result != want {
			b.Fatalf("INVALID: result_we_got %v, result_we_want %v", result, want)
		}
	}
}

func TestService_FilterPayments(t *testing.T) {
	s := newTestService()
	acc1, payments1, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, _, _, err = s.addAccount(defaultTestAccount2)
	if err != nil {
		t.Error(err)
		return
	}

	filterPayment, err := s.FilterPayments(acc1.ID, 2)
	if err != nil {
		t.Error(err)
		return
	}

	want := len(payments1)
	result := len(filterPayment)
	if !reflect.DeepEqual(result, want) {
		t.Errorf("FilterPaymnets(): INVALID: result we got %v, result we want %v", result, want)
		return
	}
}

func TestService_FilterPayments_notSuccess(t *testing.T) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = s.FilterPayments(s.nextAccountID+1, 0)
	if err == nil {
		t.Error(err)
		return
	}
}

func BenchmarkFilterPayments(b *testing.B) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		b.Error(err)
		return
	}

	acc2, payments2, _, err := s.addAccount(defaultTestAccount2)
	if err != nil {
		b.Error(err)
		return
	}

	for i := 0; i < b.N; i++ {
		payment, err := s.FilterPayments(acc2.ID, 3)
		if err != nil {
			b.Error(err)
			return
		}

		want := len(payments2)
		result := len(payment)
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("FilterPayments(): INVALID: result we got %v, result we want %v", result, want)
			return
		}
	}
}

func TestService_FilterPaymentsByFn(t *testing.T) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
	}

	_, _, _, err = s.addAccount(defaultTestAccount2)
	if err != nil {
		t.Error(err)
	}

	payment, err := s.FilterPaymentsByFn(FilterCategory,0)
	if err != nil {
		t.Error(err)
	}

	want := 2
	result := len(payment)
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("FilterPayments(): INVALID: result we got %v, result we want %v", result, want)
		return
	}
}

func BenchmarkFilterPaymentsByFn(b *testing.B) {
	s := newTestService()
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		b.Error(err)
	}

	_, _, _, err = s.addAccount(defaultTestAccount2)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		payment, err := s.FilterPaymentsByFn(FilterCategory, 10)
		if err != nil {
			b.Error(err)
			return
		}

		want := 2
		result := len(payment)
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("FilterPayments(): INVALID: result we got %v, result we want %v", result, want)
			return
		}
	}
}