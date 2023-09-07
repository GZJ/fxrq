# fxrq - Foreign Exchange Rate Queryer

fxrq is a command-line tool designed to query foreign exchange rates easily. It provides quick access to currency conversion information, making it convenient for users who need up-to-date exchange rate data.

## Installation

Make sure to install the required dependencies:

```shell
go get -u github.com/ktr0731/go-fuzzyfinder
go install github.com/gzj/fxrq
```

## Usage
### non-interactive mode
```shell
fxrq --base=USD --target=CNY --amount=1
```
This command allows you to perform a single currency conversion query with specified parameters:

- `--base`: Specify the base currency (e.g., USD).
- `--target`: Specify the target currency (e.g., CNY).
- `--amount`: Specify the amount to convert (e.g., 1).

### interactive mode
```shell
fxrq
fxrq --config=fxrq.config
fxrq --curcodefile=currencies.csv
fxrq --proxyurl="http://127.0.0.1:1080"
fxrq --apikey="*******"
fxrq --endpoint="exchangerate.host"
```

When you enter the interactive mode in `fxrq`, you'll have access to a powerful and flexible currency selection process similar to using `fzf`. Here's how it works:

1. In the first step, you'll be prompted to select the `base` currency. Choose your base currency from the list.
2. In the second step, use `ctrl + n` (next) and `ctrl + p` (previous) to navigate through the list of target currencies. Press `tab` to select the currencies you want for your query. This combination of key commands allows you to efficiently choose your target currencies without the need for complex interactions.
3. Finally, after selecting the base and target currencies, you'll be prompted to enter the `amount` for your currency conversion. Simply type in the desired amount.

Once you've completed these steps, `fxrq` will provide you with the exchange rates for your selected currencies and amount. This interactive mode offers a convenient and efficient way to perform currency conversions.

#### Additional Options

You can further customize `fxrq` with the following options:

- `--config`: Use a configuration file (e.g., fxrq.config) to set default values for base currency, target currency, and amount. The tool looks for configuration files in the following locations and uses the first one found:

  - `./.fxrq.json`
  - `~/.fxrq.json`
  - `~/.config/fxrq/fxrq.json`
  
  If any of these configuration files exist, they will be used to set default values.
- `--curcodefile`: Specify a CSV file (e.g., `currencies.csv`) that contains currency code mappings for precise currency rate queries.
- `--proxyurl`: Specify a proxy server URL if you need to access external data sources through a proxy.

- `--endpoint`: Specify the API endpoint (e.g., `--endpoint=exchangerate.host`) to use for foreign exchange rate data. Currently, `fxrq` only supports the `exchangerate.host` API as the endpoint. Please provide this option if you want to specify a different endpoint, although the default is `exchangerate.host`.
- `--apikey`: Provide your API key if required by the external data source.
