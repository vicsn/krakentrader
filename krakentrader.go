package main

import (
  "encoding/json"
	"log"
  "io/ioutil"
  "flag"
  "fmt"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/vicsn/kraken-go-api-client"
)

type Config struct {
    MaxTrade        uint64              `json:"maxTrade"`
    Public          string              `json:"public"`
    Private         string              `json:"private"`
    Users           []string            `json:"users"`
    TradingSchedule map[string][]string `json:"tradingSchedule"`
}

func main() {

  // Construct log file path based on year, month and recipient
  currentYear := strconv.Itoa(time.Now().Year())
  currentMonth := time.Now().Month().String()
  logDir := "logs/"
  logPath := logDir + currentYear + "_" + currentMonth + "_"
  _ = os.Mkdir(logDir, 0700)

  // Read config file
  configPtr := flag.String("config", "config.json", "a string")
  flag.Parse()

  config_file, err := os.Open(*configPtr)
  if err != nil {
		log.Fatal(err)
  }

  byteValue, _ := ioutil.ReadAll(config_file)
  var config Config
  json.Unmarshal(byteValue, &config)

  // Print config values
  fmt.Printf("Users: %v\n", config.Users)
  tradingSchedule := config.TradingSchedule[currentMonth]
  fmt.Printf("To trade: %v\n", tradingSchedule)

  // Read recipient from stdin
  fmt.Print("For who are you trading? For: ")
  var recipient string
  fmt.Scanln(&recipient)

  // Add recipient to log file name if we recognize the person
  for _, val := range config.Users {
    if recipient == val {
      logPath += recipient
      logPath += ".log"
      break
    }
  }

  // If the recipient wasn't recognized, abort
  if logPath[len(logPath)-4:] != ".log" {
    fmt.Printf("I don't know that person! I only know the following users indicated in the config file: %v", config.Users)
    return
  }

  // Check if log file exists for this month. As this program is only supposed
  // to run once a month, we exit if we find a log file for this month
  if _, err := os.Stat(logPath); err == nil {
    fmt.Println("You already traded this month!")
    return
  }

  // Read private key from stdin if it is not in the config file
  if config.Private == "" {
    fmt.Print("Enter private api key: ")
    fmt.Print("\033[8m") // hide user input
    fmt.Scanln(&config.Private)
    fmt.Print("\033[28m") // show user input
  }

  // Create new log file
  f, err := os.Create(logPath)
  if err != nil {
		log.Fatal(err)
  }
  defer f.Close()

  // Create api object
	api := krakenapi.New(config.Public, config.Private)

  // Get trading limits
  fmt.Println("Getting trading limits")
  var data map[string]string
  assetPairs, err := api.Query("AssetPairs", data)

	if err != nil {
		log.Fatal(err)
	}

  // Define variable to map trading pairs and minimal orders
  pairs := make(map[string]float64)

  pairsConcatenated := ""
  tradingCurrencySymbol := tradingSchedule[0][len(tradingSchedule[0])-3:]
  for _, v := range config.TradingSchedule[currentMonth] {
    // Check if the user is using a single currency to sell
    if tradingCurrencySymbol != v[len(v)-3:] {
      fmt.Println("You cannot sell different currencies, as the maxTrade config value is denoted in a single currency. Please adjust your tradingSchedule.")
      fmt.Printf("Expected: %s, Got: %s\n", tradingCurrencySymbol, v[len(v)-3:])
      return
    }

    pair := assetPairs.(map[string]interface{})[v]
    ordermin := pair.(map[string]interface{})["ordermin"]

    ordermin_float, err := strconv.ParseFloat(ordermin.(string), 64)
    if err != nil {
        log.Fatal(err)
    }

    // Map trading pairs and minimal order
    pairs[v] = ordermin_float
    pairsConcatenated = pairsConcatenated + v + ","
  }
  pairsConcatenated = strings.TrimRight(pairsConcatenated, ",")

  // Query Ticker
  fmt.Println("Getting Ticker")
  result, err := api.Query("Ticker", map[string]string{
		"pair": pairsConcatenated,
	})

	if err != nil {
		log.Fatal(err)
	}

  f.WriteString(time.Now().String())
  f.WriteString("\nTicker:")
  result_printable, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
      log.Fatal(err)
  }
  _, err_print := f.WriteString(string(result_printable))
  if err_print != nil {
		log.Fatal(err_print)
  }

  map_result := result.(map[string]interface {})
  orders := make(map[string][2]float64)
  var totalTrade uint64 = 0

  // For each of the pairs for which we want to place a trade
  for k, v := range pairs {
    // Get the current lowest ask and highest bid
    value := map_result[k].(map[string]interface {})
    ask := value["a"].([]interface{})
    bid := value["b"].([]interface{})
    ask_num := ask[0].(string)
    bid_num := bid[0].(string)

    ask_float, err := strconv.ParseFloat(ask_num, 64)
    if err != nil {
        log.Fatal(err)
    }
    bid_float, err := strconv.ParseFloat(bid_num, 64)
    if err != nil {
        log.Fatal(err)
    }

    // Calculate the price to pay, which is halfway between the current lowest
    // ask and highest bid
    price := (ask_float+bid_float)/2.0

    // save the price, amount, and price*amount we're trading
    arr := [2]float64{price, v}
    orders[k] = arr
    totalTrade += uint64(price*v)
	}

  f.WriteString("Orders:")
  orders_printable, err := json.MarshalIndent(orders, "", "  ")
  if err != nil {
      log.Fatal(err)
  }
  _, err_print2 := f.WriteString(string(orders_printable))
  if err_print2 != nil {
		log.Fatal(err_print2)
  }

  fmt.Printf("Adding orders for: %v\n", orders)

  if (totalTrade > config.MaxTrade) {
    fmt.Printf("Aborting the trades. The sum of orders would cost more than the indicated MaxTrade of: %d\n", config.MaxTrade)
    fmt.Printf("The sum of orders would cost: %d\n", totalTrade)
    return
  }

  // Ask for confirmation
  fmt.Println("Are you sure? (y/n): ")
  var confirmation string
  fmt.Scanln(&confirmation)
  if confirmation != "y" {
    fmt.Println("Alrighty, no problem. Bye!")
    return
  }

  // Execute trading orders
  for k, v := range orders {
    price := make(map[string]string)
    price["price"] = strconv.FormatFloat(v[0], 'f', 1, 64)
    order_result, err := api.AddOrder(k, "buy", "limit", strconv.FormatFloat(v[1], 'f', -1, 64), price)

    if err != nil {
      log.Fatal(err)
    }
    f.WriteString("Order result:")
    order_result_printable, err := json.MarshalIndent(order_result, "", "  ")
    if err != nil {
        log.Fatal(err)
    }
    _, err_print3 := f.WriteString(string(order_result_printable))
    if err_print3 != nil {
      log.Fatal(err_print3)
    }
  }
}
