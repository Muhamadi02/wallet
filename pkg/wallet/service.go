package wallet

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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

// ExportToFile - экспортирует аккаунты в файл.
func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func ()  {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	data := make([]byte, 0)
	lastStr := ""
	for _, account := range s.accounts {
		text := []byte(
			strconv.FormatInt(int64(account.ID), 10) + string(";") + 
			string(account.Phone) + string(";") + 
			strconv.FormatInt(int64(account.Balance), 10) + string("|"))
		data = append(data, text...)
		str := string(data)
		lastStr = strings.TrimSuffix(str, "|")
	}

	_, err = file.Write([]byte(lastStr))
	if err != nil {
		log.Print(err)
		return err
	}
	log.Printf("%#v", file)
	return nil
}

// ImportFromFile - импортирует аккаунты из файла.
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func ()  {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	content := make([]byte, 0)
	buf := make([]byte, 4)
	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}

		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buf[:read]...)
	}

	data := string(content)

	acc := strings.Split(data, "|")
	for _, tempAcc := range acc {
		tempAccount := strings.Split(tempAcc, ";")
		id, _ := strconv.ParseInt(tempAccount[0], 10, 64)

		phone := types.Phone(tempAccount[1])

		balance, _ := strconv.ParseInt(tempAccount[2], 10, 64)

		account := &types.Account{
			ID: id,
			Phone: phone,
			Balance: types.Money(balance),
		}

		s.accounts = append(s.accounts, account)
	}
	
	return nil
}

// Export - экпортирует в файл данные аккаунта платежей и избранных
func (s *Service) Export(dir string) error {
	path, _ := filepath.Abs(dir)
	os.MkdirAll(dir, 0666)

	if s.accounts != nil {
		data := make([]byte, 0)

		for _, acc := range s.accounts {
			text := []byte(
				strconv.FormatInt(int64(acc.ID), 10) + ";" + 
				string(acc.Phone) + ";" + 
				strconv.FormatInt(int64(acc.Balance), 10) + "\n")

			data = append(data, text...)
		}

		err := os.WriteFile(path + "/accounts.dump", data, 0666)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	if s.payments != nil {
		data := make([]byte, 0)

		for _, payment := range s.payments {
			text := []byte(
				string(payment.ID) + ";" + 
				strconv.FormatInt(int64(payment.AccountID), 10) + ";" + 
				strconv.FormatInt(int64(payment.Amount), 10) + ";" +
				string(payment.Category) + ";" + 
				string(payment.Status) + "\n")

			data = append(data, text...)
		}

		err := os.WriteFile(path + "/payments.dump", data, 0666)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	if s.favorites != nil {
		data := make([]byte, 0)

		for _, favorite := range s.favorites {
			text := []byte(
				string(favorite.ID) + ";" + 
				strconv.FormatInt(int64(favorite.AccountID), 10) + ";" + 
				string(favorite.Name) + ";" +
				strconv.FormatInt(int64(favorite.Amount), 10) + ";" +
				string(favorite.Categoty) + "\n")
			
			data = append(data, text...)
		}

		err := os.WriteFile(path + "/favorites.dump", data, 0666)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}

// Import - импортирует из файла данные об аккаунтах, платежей и избранных если они есть.
func (s *Service) Import(dir string) error {
	var path string
	if filepath.IsAbs(path) {
		path = filepath.Dir(dir)
	}else {
		path = dir
	}

	// import accounts
	accFile, err1 := os.ReadFile(path + "/accounts.dump")
	if err1 == nil {

		accData := string(accFile)
		accData = strings.TrimSpace(accData)

		accSlice := strings.Split(accData, "\n")
		log.Print("accounts : ", accSlice)

		for _, accImp := range accSlice {
			if len(accImp) == 0 {
				break
			}
			accStr := strings.Split(accImp, ";")
			log.Print("accStr", accStr)

			id, _ := strconv.ParseInt(accStr[0], 10, 64)
			phone := types.Phone(accStr[1])
			balance, _ := strconv.ParseInt(accStr[2], 10, 64)

			accFind, _ := s.FindAccountByID(id)
			if accFind != nil {
				accFind.Phone = phone
				accFind.Balance = types.Money(balance)
			}else {
				s.nextAccountID++
				account := &types.Account{
					ID: id,
					Phone: phone,
					Balance: types.Money(balance),
				}
				s.accounts = append(s.accounts, account)
				log.Print(account)
			}
		} 
	}else {
		log.Print(err1)
	}

	// import payments
	payFile, err2 := os.ReadFile(path + "/payments.dump")
	if err2 == nil {
		payData := string(payFile)
		payData = strings.TrimSpace(payData)

		paySlice := strings.Split(payData, "\n")
		log.Print("paySlice : ", paySlice)

		for _, payImp := range paySlice {
			if len(payImp) == 0 {
				break
			}
			payStr := strings.Split(payImp, ";")
			log.Print("payStr : ", payStr)

			id := payStr[0]
			accountID, _ := strconv.ParseInt(payStr[1], 10, 64)
			amount, _ := strconv.ParseInt(payStr[2], 10, 64)
			category := types.PaymentCategory(payStr[3])
			status := types.PaymentStatus(payStr[4])

			payAcc, _ := s.FindPaymentById(id)
			if payAcc != nil {
				payAcc.AccountID = accountID
				payAcc.Amount = types.Money(amount)
				payAcc.Category = category
				payAcc.Status = status
			} else {
				payment := &types.Payment{
					ID: id,
					AccountID: accountID,
					Amount: types.Money(amount),
					Category: category,
					Status: status,
				}
				s.payments = append(s.payments, payment)
				log.Print(payment)
			}
		}
	}else {
		log.Print(err2)
	}
	
	// import favorites
	favFile, err3 := os.ReadFile(path + "/favorites.dump")
	if err3 == nil {

		favData := string(favFile)
		favData = strings.TrimSpace(favData)

		favSlice := strings.Split(favData, "\n")
		log.Print("favSlice : ", favSlice)

		for _, favOperation := range favSlice {

			if len(favOperation) == 0 {
				break
			}
			favStr := strings.Split(favOperation, ";")
			log.Println("favStr:", favStr)

			id := favStr[0]
			accountID, _ := strconv.ParseInt(favStr[1], 10, 64)
			name := favStr[2]
			amount, _ := strconv.ParseInt(favStr[3], 10, 64)
			category := types.PaymentCategory(favStr[4])
			
			favAcc, _ := s.FindFavoriteByID(id)
			if favAcc != nil {
				favAcc.AccountID = accountID
				favAcc.Name = name
				favAcc.Amount = types.Money(amount)
				favAcc.Categoty = category
			} else {
				favorite := &types.Favorite{
					ID:        id,
					AccountID: accountID,
					Name:      name,
					Amount:    types.Money(amount),
					Categoty:  category,
				}
				s.favorites = append(s.favorites, favorite)
				log.Print(favorite)
			}
		}
	} else {
		log.Println(err3)
	}

	return nil
}

// ExportAccountHistory - выводить все платежи конкретного аккаунта
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	
	_, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	payments := []types.Payment{}
	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			payments = append(payments, *payment)
		}
	}

	if len(payments) <= 0 || payments == nil {
		return nil, ErrPaymentNotFound
	}

	return payments, nil
}

// HistoryToFiles - сохраняеть результаты функции ExportAccountHistory в файл.
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {

	_, cerr := os.Stat(dir)
	if os.IsNotExist(cerr) {
		cerr = os.Mkdir(dir, 0777)
	}
	if cerr != nil {
		return cerr
	}

	if payments == nil {
		return nil
	}

	data := make([]byte, 0)

	if len(payments) > 0 && len(payments) <= records {
		for _, payment := range payments {
			text := []byte(
				string(payment.ID) + ";" + 
				strconv.FormatInt(int64(payment.AccountID), 10) + ";" +
				strconv.FormatInt(int64(payment.Amount), 10) + ";" +
				string(payment.Category) + ";" +
				string(payment.Status) + "\n")

			data = append(data, text...)
		}

		path := dir + "/payments.dump"
		err := os.WriteFile(path, data, 0777)
		if err != nil {
			log.Print(err)
			return err
		}
	} else {
		for i, payment := range payments {

			text := []byte(
				string(payment.ID) + ";" +
					strconv.FormatInt(int64(payment.AccountID), 10) + ";" +
					strconv.FormatInt(int64(payment.Amount), 10) + ";" +
					string(payment.Category) + ";" +
					string(payment.Status) + "\n")

			data = append(data, text...)

			if (i+1) % records == 0 || i == len(payments)-1 {
				path := dir + "/payments" + strconv.Itoa((i/records)+1) + ".dump"
				err := os.WriteFile(path, data, 0777)
				if err != nil {
					log.Print(err)
					return err
				}
				data = nil
			}
		}
	}

	return nil
}

// SumPayments - суммирует платежи с помощью горутин
func (s *Service) SumPayments(goroutines int) types.Money {

	if goroutines < 1 {
		goroutines = 1
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	num := len(s.payments)/goroutines + 1
	sum := types.Money(0)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		total := types.Money(0)

		go func (val int)  {
			defer wg.Done()
			lowIndex := val * num
			highIndex := (val * num) + num

			for j := lowIndex; j < highIndex; j++ {
				if j > len(s.payments) - 1 {
					break
				}
				total += s.payments[j].Amount
			}
			mu.Lock()
			defer mu.Unlock()
			sum += total
		}(i)
	}
	
	wg.Wait()
	return sum
}

// FilterPayments - выводить все платежи определенного аккаунта.
func (s *Service) FilterPayments(accountID int64, goroutines int) ([]types.Payment, error) {
	
	_, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	if goroutines < 1 {
		goroutines = 1
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	num := len(s.payments)/goroutines + 1
	resPayments := []types.Payment{}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		tempPayments := []types.Payment{}

		go func (val int)  {
			defer wg.Done()
			lowIndex := val * num
			highIndex := (val * num) + num

			for j := lowIndex; j < highIndex; j++ {
				if j > len(s.payments) - 1 {
					break
				}

				if s.payments[j].AccountID == accountID {
					tempPayments = append(tempPayments, *s.payments[j])
				}
			}
			mu.Lock()
			defer mu.Unlock()
			resPayments = append(resPayments, tempPayments...) 
		}(i)
	}

	wg.Wait()
	return resPayments, nil
}

// FilterCategory ставит нужную категорию.
func FilterCategory(payment types.Payment) bool {
	return payment.Category == "auto"
}

// FilterPaymentsByFn фильтрует платежи по любим функциям.
func (s *Service) FilterPaymentsByFn(filter func(payment types.Payment)bool, goroutines int) ([]types.Payment, error) {
	if goroutines < 1 {
		goroutines = 1
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	num := len(s.payments)/goroutines + 1
	resPayments := []types.Payment{}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		tempPayments := []types.Payment{}

		go func (val int)  {
			defer wg.Done()
			lowIndex := val * num
			highIndex := (val * num) + num

			for j := lowIndex; j < highIndex; j++ {
				if j > len(s.payments) - 1 {
					break
				}

				if filter(*s.payments[j]) {
					tempPayments = append(tempPayments, *s.payments[j])
				}
			}
			mu.Lock()
			defer mu.Unlock()
			resPayments = append(resPayments, tempPayments...)
		}(i)
	}

	wg.Wait()
	return resPayments, nil
}