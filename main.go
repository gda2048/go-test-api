package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "net/smtp"
    "strings"

    "github.com/gorilla/mux"
)

type Rate struct {
    FromCurrency string `json:"FromCurrency"`
    ToCurrency string `json:"ToCurrency"`
    Value string `json:"Value"`
    UpdatedAt string `json:"UpdatedAt"`
}

type BinanceExchangeRateResponse struct {
    Minutes int `json:"mins"`
    Price string `json:"price"`
}

const BinanceExchangeRateURL string = "https://api.binance.com/api/v3/avgPrice"
const FileName string = "emails.json"
const FromEmail = "dmytrohoncharovgolang@gmail.com"
const FromEmailPassword = "lentsntlodioisno"

func GetEmailsFromFile() (*[]string, error){
    var emails []string

    emailsBytes, err := ioutil.ReadFile(FileName)
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(emailsBytes, &emails)
    if err != nil {
        return nil, err
    }
    return &emails, nil
}

func sendEmail(emailAddress string, content string)(error){
    log.Println(emailAddress, content)
    to := []string{emailAddress}

    host := "smtp.gmail.com"
    port := "587"
    address := host + ":" + port

    subject := "Subject: BTC Exchange Rate\n"
    message := []byte(subject + content)

    auth := smtp.PlainAuth("", FromEmail, FromEmailPassword, host)

    err := smtp.SendMail(address, auth, FromEmail, to, message)
    if err != nil {
        return err
    }
    return nil
}

func SaveEmailsToFile(emails []string) (error){
    emailsBytes, err := json.MarshalIndent(emails, "", " ")
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(FileName, emailsBytes, 0644)
    if err != nil {
        return err
    }
    return nil
}


func GetBinanceExchangeRate(FromCurrency string, ToCurrency string) (*BinanceExchangeRateResponse, error) {
    var url string = strings.Join([]string{BinanceExchangeRateURL, "?symbol=", FromCurrency, ToCurrency}, "")

    response, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }

    var rate BinanceExchangeRateResponse
    json.Unmarshal(responseData, &rate)

    return &rate, nil
}

func LatestRateBTCUAH(w http.ResponseWriter, r *http.Request){
    log.Println("Endpoint Hit: LatestRate")

    var fromCurrency string = "BTC"
    var toCurrency string = "UAH"

    rate, err := GetBinanceExchangeRate(fromCurrency, toCurrency)
    if err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    json.NewEncoder(w).Encode(rate.Price)
}

func SubscribeEmail(w http.ResponseWriter, r *http.Request){
    log.Println("Endpoint Hit: SubscribeEmail")

    r.ParseForm()
    email := r.Form.Get("email")

    emails, err := GetEmailsFromFile()
    if err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    for _, item := range *emails {
        if email == item {
            log.Println("Email exists")
            w.WriteHeader(http.StatusConflict)
            return
        }
    }


    *emails = append(*emails, email)

    err = SaveEmailsToFile(*emails)
    if err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    json.NewEncoder(w).Encode(email)
}

func SendEmails(w http.ResponseWriter, r *http.Request){
    log.Println("Endpoint Hit: SendEmails")

    var fromCurrency string = "BTC"
    var toCurrency string = "UAH"

    rate, err := GetBinanceExchangeRate(fromCurrency, toCurrency)
    if err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    emails, err := GetEmailsFromFile()
    if err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    var emailContent string = strings.Join([]string{"1", fromCurrency, " = ", rate.Price, toCurrency}, "")

    var erroredEmails []string
    for _, email := range *emails {
        err = sendEmail(email, emailContent)
        if err != nil {
            log.Println(err)
            erroredEmails = append(erroredEmails, email)
        }
    }

    json.NewEncoder(w).Encode(erroredEmails)
}

func handleRequests() {
    Router :=  mux.NewRouter().StrictSlash(true)

    Router.HandleFunc("/api/rate", LatestRateBTCUAH).Methods("GET")
    Router.HandleFunc("/api/subscribe", SubscribeEmail).Methods("POST")
    Router.HandleFunc("/api/sendEmails", SendEmails).Methods("POST")
    log.Println(http.ListenAndServe(":8081", Router))
}

func main() {
    handleRequests()
}
