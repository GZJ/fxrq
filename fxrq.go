package main

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ktr0731/go-fuzzyfinder"
)

// ------------------------------- config and args ------------------------------------------
type Config struct {
	BaseCurrency   string   `json:"base_currency"`
	TargetCurrency []string `json:"target_currency"`
	Amount         string
	CurCodeFile    string `json:"currency_code_file"`
	ProxyURL       string `json:"proxy_url"`
	Endpoint       string `json:"endpoint"`
	ApiKey         string `json:"api_key"`
}

func NewConfig(filePath string) (Config, error) {
	if filePath == "" {
		paths := []string{
			"./.fxrq.json",
			"~/.fxrq.json",
			"~/.config/fxrq/fxrq.json",
		}
		for _, path := range paths {
			expandedPath, err := filepath.Abs(path)
			if err == nil {
				if _, err := os.Stat(expandedPath); err == nil {
					filePath = expandedPath
					break
				}
			}
		}
		if filePath == "" {
			return Config{}, errors.New("config file path not provided")
		}
	}
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Println("Failed to decode config file:", err)
		return Config{}, err
	}

	return config, nil
}

func flags() Config {
	configFile := flag.String("config", "", "path to config file")
	baseCurrency := flag.String("base", "", "base currency")
	targetCurrency := flag.String("target", "", "target currency, use commas to separate multiple currencies")
	amount := flag.String("amount", "", "amount of base currency")
	curCodeFile := flag.String("curcodefile", "", "currency code file")
	endpoint := flag.String("endpoint", "exchangerate.host", "API endpoint")
	apiKey := flag.String("apikey", "", "API key")
	proxyUrl := flag.String("proxyurl", "", "proxy url")

	flag.Parse()

	config, err := NewConfig(*configFile)
	if err != nil {
		config = Config{
			BaseCurrency: *baseCurrency,
			Amount:       *amount,
			CurCodeFile:  *curCodeFile,
			Endpoint:     *endpoint,
			ApiKey:       *apiKey,
			ProxyURL:     *proxyUrl,
		}
		if *targetCurrency == "" {
			config.TargetCurrency = []string{}
		} else {
			config.TargetCurrency = strings.Split(*targetCurrency, ",")
		}
	} else {
		if *baseCurrency != "" {
			config.BaseCurrency = *baseCurrency
		}
		if *targetCurrency != "" {
			config.TargetCurrency = strings.Split(*targetCurrency, ",")
		}
		if *amount != "" {
			config.Amount = *amount
		}
		if *curCodeFile != "" {
			config.CurCodeFile = *curCodeFile
		}
		if *endpoint != "" {
			config.Endpoint = *endpoint
		}
		if *apiKey != "" {
			config.ApiKey = *apiKey
		}
		if *proxyUrl != "" {
			config.ProxyURL = *proxyUrl
		}
	}

	return config
}

// ------------------------------- currency table ------------------------------------------
//
//go:embed embed/currencies.csv
var CurrenciesData []byte

type Currency struct {
	Symbol string
	Iso    string
	Region string
	Name   string
}

type CurrencyList struct {
	currencies []Currency
}

func NewCurrencyList() *CurrencyList {
	return &CurrencyList{}
}

func (cl *CurrencyList) Init(filename string) {
	var (
		currenciesData []byte = CurrenciesData
		records        [][]string
		err            error
	)

	if filename != "" {
		records, err = cl.Read(filename)
		if err != nil {
			log.Println("Failed to read currencies:", err)
			os.Exit(1)
		}
	} else {
		records, err = cl.readCSV(strings.NewReader(string(currenciesData)))
		if err != nil {
			log.Println("Failed to read currencies:", err)
			os.Exit(1)
		}
	}
	cl.convert(records)
}

func (cl *CurrencyList) convert(records [][]string) {
	var currencies []Currency
	for _, record := range records {
		currency := Currency{
			Symbol: record[0],
			Iso:    record[1],
			Name:   record[2],
			Region: record[3],
		}
		currencies = append(currencies, currency)
	}

	cl.currencies = currencies
}

func (cl *CurrencyList) Read(filename string) ([][]string, error) {
	var records [][]string
	var err error

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch filepath.Ext(filename) {
	case ".csv":
		records, err = cl.readCSV(f)
	case ".json":
		records, err = cl.readJSON(f)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", filepath.Ext(filename))
	}

	if err != nil {
		return nil, err
	}

	return records, nil
}

func (cl *CurrencyList) readCSV(fr io.Reader) ([][]string, error) {
	r := csv.NewReader(fr)

	// skip first line in csv
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (cl *CurrencyList) readJSON(fr io.Reader) ([][]string, error) {
	var records [][]string
	if err := json.NewDecoder(fr).Decode(&records); err != nil {
		return nil, err
	}

	return records, nil
}

func (cl *CurrencyList) FuzzGetCode() string {
	id, err := fuzzyfinder.Find(
		cl.currencies,
		func(i int) string {
			return fmt.Sprintf("%s - %s - %s (%s)", cl.currencies[i].Region, cl.currencies[i].Name, cl.currencies[i].Iso, cl.currencies[i].Symbol)
		},
	)
	if err != nil {
		log.Println("Failed to get keyword:", err)
		os.Exit(1)
	}
	return cl.currencies[id].Iso
}

func (cl *CurrencyList) FuzzGetCodes() []string {
	ids, err := fuzzyfinder.FindMulti(
		cl.currencies,
		func(i int) string {
			return fmt.Sprintf("%s - %s - %s (%s)", cl.currencies[i].Region, cl.currencies[i].Name, cl.currencies[i].Iso, cl.currencies[i].Symbol)
		},
	)
	if err != nil {
		log.Println("Failed to get keyword:", err)
		os.Exit(1)
	}
	codes := make([]string, 0)
	for _, id := range ids {
		codes = append(codes, cl.currencies[id].Iso)
	}
	return codes
}

// ------------------------------- endpoint ------------------------------------------
type Endpoint interface {
	Query(baseCurrency string, targetCurrency []string, amount string) Result
}

func CreateEndpoints(apiKey, proxyURL string) map[string]Endpoint {
	endpoints := make(map[string]Endpoint)
	endpoints["exchangerate.host"] = NewExchangeRateHost(apiKey, proxyURL)
	return endpoints
}

type Result struct {
	Date   string
	Base   string
	Target []string
	Rates  map[string]float64
}

// ----------------------- endpoint exchangerate.host ---------------------------------
type ExchangeRateHost struct {
	Name     string
	URL      string
	APIKey   string
	ProxyURL string
}

func NewExchangeRateHost(apiKey, proxyURL string) *ExchangeRateHost {
	return &ExchangeRateHost{
		Name:     "exchangerate.host",
		URL:      "https://api.exchangerate.host/latest?base=%s&symbols=%s&amount=%s",
		APIKey:   apiKey,
		ProxyURL: proxyURL,
	}
}

func (erh *ExchangeRateHost) Query(baseCurrency string, targetCurrency []string, amount string) Result {
	var client *http.Client
	if erh.ProxyURL != "" {
		proxyURL, err := url.Parse(erh.ProxyURL)
		if err != nil {
			panic(err)
		}

		tr := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		client = &http.Client{
			Transport: tr,
		}
	} else {
		client = &http.Client{}
	}

	u := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s&symbols=%s&amount=%s", baseCurrency, strings.Join(targetCurrency, ","), amount)
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}

	return erh.convert2result(resp, baseCurrency, targetCurrency)
}

func (erh *ExchangeRateHost) convert2result(resp *http.Response, baseCurrency string, targetCurrency []string) Result {
	var (
		respm  map[string]interface{}
		result Result
	)

	json.NewDecoder(resp.Body).Decode(&respm)
	if dateStr, ok := respm["date"].(string); ok {
		result.Date = dateStr
	} else {
		log.Println("Date is not a string")
	}
	result.Base = baseCurrency
	result.Target = targetCurrency
	result.Rates = make(map[string]float64)
	if rates, ok := respm["rates"].(map[string]interface{}); ok {
		for k, v := range rates {
			if vf, ok := v.(float64); ok {
				result.Rates[k] = vf
			} else {
				log.Println("Rates is not a map[string]interface{}")
			}
		}
	} else {
		log.Println("Rate is not a string")
	}
	return result
}

// ------------------------------- main ------------------------------------------
var cfg Config

func main() {
	cfg = flags()

	cl := NewCurrencyList()
	cl.Init(cfg.CurCodeFile)

	eps := CreateEndpoints(cfg.ApiKey, cfg.ProxyURL)
	ep := eps[cfg.Endpoint]

	if len(cfg.BaseCurrency) != 0 && len(cfg.TargetCurrency) != 0 && len(cfg.Amount) != 0 {
		result := ep.Query(cfg.BaseCurrency, cfg.TargetCurrency, cfg.Amount)
		fmt.Println("Date: ", result.Date)
		fmt.Println("Base: ", result.Base)
		fmt.Println("Target: ", result.Target)
		fmt.Println("Rates: ", result.Rates)
	} else {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		go interactive(cfg, cl, ep)
		<-interrupt
	}
}

func interactive(cfg Config, cl *CurrencyList, ep Endpoint) {
	var (
		baseCurrency   string
		targetCurrency []string
	)

	if cfg.BaseCurrency == "" {
		baseCurrency = cl.FuzzGetCode()
	} else {
		baseCurrency = cfg.BaseCurrency
	}
	if len(cfg.TargetCurrency) == 0 {
		targetCurrency = cl.FuzzGetCodes()
	} else {
		targetCurrency = cfg.TargetCurrency
	}

	fmt.Println(cfg.Endpoint)
	for {
		var amount string
		fmt.Println("-------------------------")
		fmt.Print("Enter amount: ")
		_, err := fmt.Scanln(&amount)
		if err != nil {
			log.Println(err)
			continue
		}
		result := ep.Query(baseCurrency, targetCurrency, amount)
		fmt.Println("Date: ", result.Date)
		fmt.Println("Base: ", result.Base)
		fmt.Println("Target: ", result.Target)
		fmt.Println("Rates: ", result.Rates)
	}
}
