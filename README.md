Kraken Trading Command Line Application
===================

> :warning: **Use at your own risk**. Review the app (it is just over 200 LOC) and its [dependency](https://github.com/beldur/kraken-go-api-client) before use.

This app places limit orders on Kraken for the tradingpairs which you specify
in the config file, for a price halfway between the current lowest ask and
highest bid.  This makes it useful for [Dollar Cost
Averaging](https://www.investopedia.com/terms/d/dollarcostaveraging.asp).

## Setup
Create [API keys on
Kraken](https://support.kraken.com/hc/en-us/articles/360000919966-How-to-generate-an-API-key-pair-)
and save them in a safe place.

Create a `config.json` file in this repository with the following contents:
- The maximum amount you wish to trade, in the currency specified by your tradingSchedule.
- Kraken public API key.
- Kraken private API key, you may leave this empty if you want to pass it via stdin.
- List of user names for which you want to run the app.
- List of pairs which you want to trade in any particular month.

Here is an example:
```
{
  "MaxTrade": 0,
  "public":"<public key>",
  "private":"<private key>",
  "users": ["<your name>", "<other name>"],
  "tradingSchedule": {
    "January":  ["XETHZEUR"],
    "February": ["XXBTZEUR"],
    "March":    ["XETHZEUR"],
    "April":    ["XXBTZEUR"],
    "May":      ["XETHZEUR"],
    "June":     ["XXBTZEUR"],
    "July":     ["XETHZEUR"],
    "August":   ["XXBTZEUR"],
    "September":["XETHZEUR"],
    "October":  ["XXBTZEUR"],
    "November": ["XETHZEUR"],
    "December": ["XXBTZEUR"]
  }
}
```

## Usage
To build and run the app on your OS of choice:
```
./build.sh
./build/krakenapp{_mac/_linux/.exe} --config <config.json>
```

The app will ask for the following information:
- If `config.json` doesn't contain any private key, the app will ask you to pass it via stdin.  
- For which user you want to execute the trade. Only users indicated in the config file are allowed.  
- If you are sure you want to execute a set of particular trades.  

Trades are logged in the `logs` directory. If a trade for that particular user
has already been made in this month, the trades will not be executed and the
app will exit.

## TODOs
- Move TODOs into issues.
- The current price may be suboptimal: when prices keep rising the order may
  never be fullfilled. When prices keep falling you may be paying "too much".
  Therefore, a market order may be more appropriate.
- Split program into functions and check for DRYness.
- Add tests
