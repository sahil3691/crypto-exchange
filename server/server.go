package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/sahildhargave/crypto/orderbook"
)

const (
	MarketETH          Market    = "ETH"
	MarketOrder        OrderType = "MARKET"
	LimitOrder         OrderType = "LIMIT"
	exchangePrivateKey           = "7d226a8ebb7cf08281489934a922e45e286ebfb22325deb2d8110ecfaee814fb"
)

type (
	OrderType string
	Market    string

	PlaceOrderRequest struct {
		UserID int64
		Type   OrderType // limit or market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		UserID    int64
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	MatchedOrder struct {
		Price float64
		Size  float64
		ID    int64
	}
)

func StartServer() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	buyeraddressStr := "0xE87EAc47694C7e9fe37cCfdDd5Ca364E08bB29b4"
	buyerbalance, err := client.BalanceAt(context.Background(), common.HexToAddress(buyeraddressStr), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("buyer:", buyerbalance)

	selleraddressStr := "0xd6c8e9E4a9e291821734D3a0431cC633fFE9Ee68"
	sellerbalance, err := client.BalanceAt(context.Background(), common.HexToAddress(selleraddressStr), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("seller:", sellerbalance)
	pkStr8 := "732b2c89215dae2be63674b9739f5f82af38eb7d904dc7676c391f78553f3d1d"

	user8 := NewUser(pkStr8, 8)

	ex.Users[user8.ID] = user8

	pkStr7 := "201c88bfc08e5eb9d98b2ec92802488e9777a36bd705eb9afeb1a9d174c62322"

	user7 := NewUser(pkStr7, 7)
	ex.Users[user7.ID] = user7

	johnPk := "3ff52bb8346bfa5d89f74b092ed7b730ec39442221d8d78a976968b1e66b9f35"
	john := NewUser(johnPk, 666)
	ex.Users[john.ID] = john

	johnAddress := "0x7514A85343Ad5B789954f68cF156F5D09Cbc4B43"
	johnBalance, err := client.BalanceAt(context.Background(), common.HexToAddress(johnAddress), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("john:", johnBalance)

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/book/ask", ex.handleGetBook)
	e.DELETE("/order/:id", ex.cancelOrder)
    e.GET("/book/:market/bid", ex.handleGetBestBid)
	e.GET("/book/:market/ask", ex.handleGetBestAsk)
	//address := "0xACa94ef8bD5ffEE41947b4585a84BdA5a3d3DA6E"
	//balance, _ := ex.Client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	//fmt.Println(balance)

	//	client, err := ethclient.Dial("http://localhost:8545")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	ctx := context.Background()
	//	//address := common.HexToAddress("0x0468Ee9dab4F6f1eed73a732E632FfD98c205C00")
	//
	//	privateKey, err := crypto.HexToECDSA("7d226a8ebb7cf08281489934a922e45e286ebfb22325deb2d8110ecfaee814fb")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	publicKey := privateKey.Public()
	//	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	//	if !ok {
	//		log.Fatal(err)
	//	}
	//
	//	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//
	//	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	value := big.NewInt(1000000000000000000) // in wei (1 eth)
	//	gasLimit := uint64(21000)                // in units
	//	//gasPrice := big.NewInt(30000000000) // in wei (30 gwei)
	//	gasPrice, err := client.SuggestGasPrice(context.Background())
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	//	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	//
	//	chainID, err := client.NetworkID(context.Background())
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	if err := client.SendTransaction(context.Background(), signedTx); err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	balance, err := client.BalanceAt(ctx, toAddress, nil)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	fmt.Println(balance)

	e.Start(":3000")
}

type User struct {
	ID         int64
	orders     map[int64]*User
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewUser(privKey string, id int64) *User {
	pk, err := crypto.HexToECDSA(privKey)
	if err != nil {
		panic(err)
	}
	return &User{
		ID:         id,
		PrivateKey: pk,
	}
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type Exchange struct {
	Client     *ethclient.Client
	Users      map[int64]*User
	orders     map[int64]int64
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	return &Exchange{
		Client:     client,
		Users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		PrivateKey: pk,
		orderbooks: orderbooks,
	}, nil
}

type GetOrdersResponse struct{
	Ask []Order
	Bids []Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]

	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"msg": "market not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				UserID:    order.UserID,
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				UserID:    order.UserID,
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}

type PriceResponse struct {
	Price float64
}

//func (ex *Exchange) handleGetBestBid(c echo.Context) error {
//   var (
//	 market = Market(c.Param("market"))
//	 ob = ex.orderbooks[market]
//	 order = Order{}
//   )
//
//   if len(ob.Bids()) == 0 {
//	 return c.JSON(http.StatusOK, order)
//   }
//
//   bestLimit := ob.Bids()[0]
//   bestOrder := bestLimit.Orders[0]
//
//   order.Price = bestLimit.Price
//   order.UserID = bestOrder.UserID
//
//   return c.JSON(http.StatusOK, order)
//}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	if len(ob.Bids()) == 0 {
		return fmt.Errorf("the bids are empty")
	}
	bestBidPrice := ob.Bids()[0].Price
	pr := PriceResponse{
		Price: bestBidPrice,
	}

	return c.JSON(http.StatusOK, pr)
}

//func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
//  var (
//	market = Market(c.Param("market"))
//	ob     = ex.orderbooks[market]
//	order = Order{}
//  )
//
//  if len(ob.Asks()) == 0 {
//	return c.JSON(http.StatusOK, order)
//  }
//
//  bestLimit := ob.Asks()[0]
//  bestOrder := bestLimit.Orders[0]
//
//  order.Price = bestLimit.Price
//  order.UserID = bestOrder.UserID
//
//  return c.JSON(http.StatusOK, order)
//}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	if len(ob.Asks()) == 0 {
		return fmt.Errorf("the asks are empty")
	}

	bestAskPrice := ob.Asks()[0].Price
	pr := PriceResponse{
		Price: bestAskPrice,
	}
	return c.JSON(http.StatusOK, pr)
}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)
	log.Println("order canceled id =>", id)

	return c.JSON(http.StatusOK, map[string]interface{}{"msg": "order deleted"})
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}

	totalSizeFilled := 0.0
	sumPrice := 0.0
	for i := 0; i < len(matchedOrders); i++ {
		id := matches[i].Bid.ID

		if isBid {
			id = matches[i].Ask.ID
		}
		matchedOrders[i] = &MatchedOrder{
			ID:    id,
			Size:  matches[i].SizeFilled,
			Price: matches[i].Price,
		}
		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price
	}

	avgPrice := sumPrice / float64(len(matches))

	log.Printf("filled market order => %d | size: [%.2f] | avgPrice: [%.2f]", order.ID, totalSizeFilled, avgPrice)
	return matches, matchedOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)
	log.Printf("New Limit Order => type: [%t] | Price [%.2f] | Size [%.2f]", order.Bid, order.Limit.Price, order.Size)
	return nil
}

type PlaceOrderResponse struct {
	OrderID int64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}
	market := Market(placeOrderData.Market)

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)
	// Limit Order
	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
	}

	// Market Order

	if placeOrderData.Type == MarketOrder {
		matches, _ := ex.handlePlaceMarketOrder(market, order)
		if err := ex.handleMatches(matches); err != nil {
			return err
		}
		//return c.JSON(http.StatusOK, map[string]any{"matches": matchedOrders})
	}

	resp := &PlaceOrderResponse{
		OrderID: order.ID,
	}

	return c.JSON(200, resp)
}
 
func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("User Not Found: %d", match.Ask.UserID)
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("User Not Found: %d", match.Bid.UserID)
		}

		toAddress := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		// this is only used for the fees
		// exchangePubKey := ex.PrivateKey.Public()
		// publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
		// if !ok {
		// return fmt.Errorf("error casting public key to ECDSA")
		//}

		amount := big.NewInt(int64(match.SizeFilled))

		transferETH(ex.Client, fromUser.PrivateKey, toAddress, amount)

	}
	return nil
}

func transferETH(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	return client.SendTransaction(ctx, signedTx)
}
