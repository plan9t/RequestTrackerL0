package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func GetOrdersFromPostgreSQL() ([]Order, error) {
	// Подключение к PostgreSQL
	dbURL := "postgresql://plan9t:plan9t@localhost:5432/WB?sslmode=disable"

	// Открытие соединения с базой данных
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
SELECT
    o.order_uid,
    o.track_number,
    o.entry,
    o.locale,
    o.internal_signature,
    o.customer_id,
    o.delivery_service,
    o.shardkey,
    o.sm_id,
    o.date_created,
    o.oof_shard,
    d.name AS delivery_name,
    d.phone AS delivery_phone,
    d.zip AS delivery_zip,
    d.city AS delivery_city,
    d.address AS delivery_address,
    d.region AS delivery_region,
    d.email AS delivery_email,
    p.transaction,
    p.request_id,
    p.currency,
    p.provider,
    p.amount,
    p.payment_dt,
    p.bank,
    p.delivery_cost,
    p.goods_total,
    p.custom_fee,
    i.chrt_id,
    i.track_number AS item_track_number,
    i.price AS item_price,
    i.rid AS item_rid,
    i.name AS item_name,
    i.sale AS item_sale,
    i.size AS item_size,
    i.total_price AS item_total_price,
    i.nm_id AS item_nm_id,
    i.brand AS item_brand,
    i.status AS item_status
FROM
    orders o
JOIN
    delivery d ON o.pk_order_id = d.pk_order_id
JOIN
    payment p ON o.pk_order_id = p.pk_order_id
JOIN
    items i ON o.pk_order_id = i.pk_order_id
    ORDER BY o.pk_order_id DESC
LIMIT 50;
`)
	if err != nil {
		fmt.Println("Ошибка при выполнении запроса к базе данных:", err)
		return nil, nil
	}
	defer rows.Close()

	var (
		currentOrderUID string
		currentOrder    Order
	)

	var orders []Order

	//counter := 1

	for rows.Next() {

		//fmt.Println("Counter rows = ", counter)
		//counter++

		var (
			orderUID       string
			trackNumber    string
			entry          string
			locale         string
			internalSig    string
			customerID     string
			deliveryServ   string
			shardKey       string
			smID           int
			dateCreated    time.Time
			oofShard       string
			deliveryName   string
			deliveryPhone  string
			deliveryZip    string
			deliveryCity   string
			deliveryAddr   string
			deliveryRegion string
			deliveryEmail  string
			trans          string
			reqID          string
			currency       string
			provider       string
			amount         int
			paymentDt      int
			bank           string
			deliveryCost   int
			goodsTotal     int
			customFee      int
			chrtID         int
			itemTrackNum   string
			price          int
			rid            string
			name           string
			sale           int
			size           string
			totalPrice     int
			nmID           int
			brand          string
			status         int
		)

		err := rows.Scan(
			&orderUID, &trackNumber, &entry, &locale, &internalSig, &customerID, &deliveryServ, &shardKey, &smID, &dateCreated, &oofShard,
			&deliveryName, &deliveryPhone, &deliveryZip, &deliveryCity, &deliveryAddr, &deliveryRegion, &deliveryEmail,
			&trans, &reqID, &currency, &provider, &amount, &paymentDt, &bank, &deliveryCost, &goodsTotal, &customFee,
			&chrtID, &itemTrackNum, &price, &rid, &name, &sale, &size, &totalPrice, &nmID, &brand, &status,
		)
		if err != nil {
			fmt.Println("Ошибка при сканировании результата:", err)
			return nil, nil
		}

		//fmt.Println("currentOrderUID = " + currentOrderUID)
		//fmt.Println("orderUID = " + orderUID)

		// Если текущий orderUID отличается от предыдущего, создаем новый элемент Order
		if currentOrderUID != orderUID {

			orders = append(orders, currentOrder)

			currentOrderUID = orderUID

			// Создаем новый заказ
			currentOrder = Order{
				OrderUID:          orderUID,
				TrackNumber:       trackNumber,
				Entry:             entry,
				Locale:            locale,
				InternalSignature: internalSig,
				CustomerID:        customerID,
				DeliveryService:   deliveryServ,
				ShardKey:          shardKey,
				SmID:              smID,
				DateCreated:       dateCreated,
				OofShard:          oofShard,
				Delivery: Delivery{
					Name:    deliveryName,
					Phone:   deliveryPhone,
					Zip:     deliveryZip,
					City:    deliveryCity,
					Address: deliveryAddr,
					Region:  deliveryRegion,
					Email:   deliveryEmail,
				},
				Payment: Payment{
					Transaction:  trans,
					RequestID:    reqID,
					Currency:     currency,
					Provider:     provider,
					Amount:       amount,
					PaymentDt:    paymentDt,
					Bank:         bank,
					DeliveryCost: deliveryCost,
					GoodsTotal:   goodsTotal,
					CustomFee:    customFee,
				},
				Items: []Item{},
			}

			item := Item{
				ChrtID:      chrtID,
				TrackNumber: itemTrackNum,
				Price:       price,
				RID:         rid,
				Name:        name,
				Sale:        sale,
				Size:        size,
				TotalPrice:  totalPrice,
				NmID:        nmID,
				Brand:       brand,
				Status:      status,
			}
			currentOrder.Items = append(currentOrder.Items, item)

		} else {

			item := Item{
				ChrtID:      chrtID,
				TrackNumber: itemTrackNum,
				Price:       price,
				RID:         rid,
				Name:        name,
				Sale:        sale,
				Size:        size,
				TotalPrice:  totalPrice,
				NmID:        nmID,
				Brand:       brand,
				Status:      status,
			}
			currentOrder.Items = append(currentOrder.Items, item)
			// fmt.Println("ОТРАБОТАЛ ПУШ ИТЕМА В ИТЕМС")
		}

		// пуш в итоговый массив с ордерами если 1. Отличается currentOrderUID и OrderUID; 2. Последняя строка.
		//orders = append(orders, currentOrder)

		// Вывод текущего ордера

		// fmt.Println(currentOrder)
	}
	orders = append(orders, currentOrder)

	if len(orders) > 0 {
		orders = orders[1:]
	}

	// fmt.Println("ИТОГОВЫЙ МАССИВ СТРУКТУР:")
	// fmt.Println(orders)
	return orders, nil
}

// parseJSON функция
func parseJSON(jsonData []byte) (Order, error) {
	var orderData Order
	err := json.Unmarshal(jsonData, &orderData)
	return orderData, err
}

// Сохранение данных в PostgreSQL
func SaveToPostgreSQL(orderData Order) error {

	// Подключение к PostgreSQL
	dbURL := "postgresql://plan9t:plan9t@localhost:5432/WB?sslmode=disable"

	// Открытие соединения с базой данных
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Создание транзакции
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Сохранение данных в таблицы базы данных
	// Сохранение данных в таблицу "orders"
	_, err = tx.Exec(
		`
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING pk_order_id`,
		orderData.OrderUID,
		orderData.TrackNumber,
		orderData.Entry,
		orderData.Locale,
		orderData.InternalSignature,
		orderData.CustomerID,
		orderData.DeliveryService,
		orderData.ShardKey,
		orderData.SmID,
		orderData.DateCreated,
		orderData.OofShard,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Получение pk_order_id для вставленной записи
	var pkOrderID int
	err = tx.QueryRow("SELECT pk_order_id FROM orders WHERE order_uid = $1", orderData.OrderUID).Scan(&pkOrderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Сохранение данных в таблицу "delivery"
	_, err = tx.Exec(`
		INSERT INTO delivery (pk_order_id, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		pkOrderID, orderData.Delivery.Name, orderData.Delivery.Phone, orderData.Delivery.Zip, orderData.Delivery.City,
		orderData.Delivery.Address, orderData.Delivery.Region, orderData.Delivery.Email)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Сохранение данных в таблицу "payment"
	_, err = tx.Exec(`
	INSERT INTO payment (pk_order_id, transaction, request_id, currency, provider, amount, payment_dt, bank, 
		delivery_cost, goods_total, custom_fee)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		pkOrderID,
		orderData.Payment.Transaction,
		orderData.Payment.RequestID,
		orderData.Payment.Currency,
		orderData.Payment.Provider,
		orderData.Payment.Amount,
		orderData.Payment.PaymentDt,
		orderData.Payment.Bank,
		orderData.Payment.DeliveryCost,
		orderData.Payment.GoodsTotal,
		orderData.Payment.CustomFee,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Сохранение данных в таблицу items (через цикл т.к в JSON-е получаем массив с итемами)
	for _, item := range orderData.Items {
		_, err = tx.Exec(`
		INSERT INTO items (pk_order_id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			pkOrderID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.RID,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Завершение транзакции
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Println("Данные успешно сохранены в PostgreSQL!")
	return nil
}
