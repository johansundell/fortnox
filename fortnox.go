// fortnox project fortnox.go
package fortnox

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type AuthInfo struct {
	AccessToken  string
	ClientSecret string
	AuthCode     string
}

type Connection struct {
	AuthInfo
}

type authTokenResp struct {
	Authorization authorization
}

type authorization struct {
	AccessToken string
}

type Customer struct {
	Name               string
	OrganisationNumber string
	CustomerNumber     string
	Url                string `json:"@url"`
	Address1           string
	Email              string
	OurReference       string
	YourReference      string
	ZipCode            string
	City               string
}

type customerSearch struct {
	Customers []Customer
}

type InvoiceRow struct {
	ArticleNumber     string
	DeliveredQuantity string
}

type Invoice struct {
	CustomerNumber string
	InvoiceRows    []InvoiceRow
	DocumentNumber string
}

type InvoiceResult struct {
	Invoice Invoice
}

// TODO: Rename me
type Article struct {
	ArticleNumber string
	Description   string
}

type ArticleResult struct {
	Article Article
}

func GetAuthToken(AuthCode, ClientSecret string) (string, error) {
	conn := Connection{AuthInfo{"", ClientSecret, AuthCode}}
	result, err := conn.getHttpClientBody("", nil)
	if err != nil {
		return "", err
	}
	r := new(authTokenResp)
	if err := json.Unmarshal(result, r); err != nil {
		return "", err
	}
	return r.Authorization.AccessToken, nil
}

func NewConnection(AccessToken, ClientSecret string) Connection {
	return Connection{AuthInfo{AccessToken, ClientSecret, ""}}
}

func (conn *Connection) GetCustomerByOrgNr(orgnr string) (Customer, error) {
	c := Customer{}
	v := url.Values{}
	v.Add("organisationnumber", orgnr)
	b, err := conn.getHttpClientBody("customers/?"+v.Encode(), nil)
	if err != nil {
		return Customer{}, err
	}
	r := new(customerSearch)
	if err := json.Unmarshal(b, r); err != nil {
		return Customer{}, err
	}
	if len(r.Customers) == 0 {
		return c, errors.New("Could not find customer")
	}
	c = r.Customers[0]
	return c, nil
}

func (conn *Connection) CreateCustomer(c Customer) (Customer, error) {

	if _, err := conn.postData(response{"Customer": c}.Bytes(), "customers"); err != nil {
		return Customer{}, err
	}

	// TODO: Maybe not the best way ;)
	return conn.GetCustomerByOrgNr(c.OrganisationNumber)
}

type response map[string]interface{}

func (r response) String() (s string) {
	s = string(r.Bytes())
	return
}

func (r response) Bytes() (b []byte) {
	b, _ = json.Marshal(r)
	return
}

func (conn *Connection) getHttpClientBody(part string, data map[string]string) ([]byte, error) {
	return conn.postDataWithMethod(nil, part, "GET", data)
}

// TODO: Rebuild this using the url and make it work with all the diffrent types
func (conn *Connection) UpdateCustomer(c Customer) (Customer, error) {
	if _, err := conn.postDataWithMethod(response{"Customer": c}.Bytes(), "customers/"+c.CustomerNumber, "PUT", nil); err != nil {
		return Customer{}, err
	}

	// TODO: Maybe not the best way ;)
	return conn.GetCustomerByOrgNr(c.OrganisationNumber)
}

func (conn *Connection) postData(data []byte, part string) ([]byte, error) {
	return conn.postDataWithMethod(data, part, "POST", nil)
}

func (conn *Connection) postDataWithMethod(data []byte, part string, method string, headData map[string]string) ([]byte, error) {
	client := &http.Client{}
	url := "https://api.fortnox.se/3/" + part
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if len(conn.AccessToken) != 0 {
		req.Header.Add("Access-Token", conn.AccessToken)
	}
	if len(conn.ClientSecret) != 0 {
		req.Header.Add("Client-Secret", conn.ClientSecret)
	}
	if len(conn.AuthCode) != 0 {
		req.Header.Add("Authorization-Code", conn.AuthCode)
	}

	for key, value := range headData {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (conn *Connection) CreateInvoice(inv Invoice) (Invoice, error) {
	b, _ := conn.postData(response{"Invoice": inv}.Bytes(), "invoices")
	r := new(InvoiceResult)
	if err := json.Unmarshal(b, r); err != nil {
		return Invoice{}, err
	}
	return r.Invoice, nil
}

func (conn *Connection) GetArticle(no string) (Article, error) {
	result, err := conn.getHttpClientBody("articles/"+no, nil)
	if err != nil {
		return Article{}, err
	}
	art := new(ArticleResult)
	if err := json.Unmarshal(result, art); err != nil {
		return Article{}, err
	}
	return art.Article, nil
}
