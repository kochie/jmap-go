package jmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	CORE     = "urn:ietf:params:jmap:core"
	MAIL     = "urn:ietf:params:jmap:mail"
	CONTACTS = "urn:ietf:params:jmap:contacts"
)

const (
	MethodEcho       = "Core/echo"
	MethodGetContact = "Contact/get"
	MethodSetContact = "Contact/set"
)

const URL = "https://api.fastmail.com/.well-known/jmap"
const TOKEN = "fmb1-cb4140a6-c8a9bb1770306f2085bb1af7c0b70742-1657184400-9e770cdce6c33bdcba76557bde355152"

//u=cb4140a6

type Id string

type Error struct {
	Type   string `json:"type"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

type File struct {
	BlobId Id     `json:"blobId"`
	Type   string `json:"type,omitempty"`
	Name   string `json:"name,omitempty"`
	Size   uint   `json:"size,omitempty"`
}

type ContactInformation struct {
	Type      string `json:"type"`
	Label     string `json:"label,omitempty"`
	Value     string `json:"value"`
	IsDefault bool   `json:"isDefault"`
}

type Address struct {
	Type      string `json:"type"`
	Label     string `json:"label,omitempty"`
	Street    string `json:"street"`
	Locality  string `json:"locality"`
	Region    string `json:"region"`
	Postcode  string `json:"postcode"`
	Country   string `json:"country"`
	IsDefault bool   `json:"isDefault"`
}

type Contact struct {
	Id          Id
	IsFlagged   bool                 `json,xml:"isFlagged"`
	Avatar      *File                `json:"avatar,omitempty"`
	Prefix      string               `json:"prefix"`
	FirstName   string               `json:"firstName"`
	LastName    string               `json:"lastName"`
	Suffix      string               `json:"suffix"`
	Nickname    string               `json:"nickname"`
	Birthday    string               `json:"birthday"`
	Anniversary string               `json:"anniversary"`
	Company     string               `json:"company"`
	Department  string               `json:"department"`
	JobTitle    string               `json:"jobTitle"`
	Emails      []ContactInformation `json:"emails"`
	Phones      []ContactInformation `json:"phones"`
	Online      []ContactInformation `json:"online"`
	Addresses   []Address            `json:"addresses"`
	Notes       string               `json:"notes"`
}

type GetRequestArguments struct {
	AccountId  Id       `json:"accountId"`
	Ids        []Id     `json:"ids,omitempty"`
	Properties []string `json:"properties,omitempty"`
}

type ObjectConstraint interface {
	Contact
}

type GetResponseArguments struct {
	AccountId Id                       `json:"accountId"`
	State     string                   `json:"state"`
	List      []map[string]interface{} `json:"list"`
	NotFound  []Id                     `json:"notFound"`
}

type PatchObject map[string]interface{}

type SetRequestArguments struct {
	AccountId Id                  `json:"accountId"`
	IfInState *string             `json:"ifInState,omitempty"`
	Create    *map[Id]interface{} `json:"create,omitempty"`
	Update    *map[Id]PatchObject `json:"update,omitempty"`
	Destroy   *[]Id               `json:"destroy"`
}

type Arguments interface {
	ResponseArguments | RequestArguments
}

type EchoRequestArguments interface{}

type RequestArguments interface {
	GetRequestArguments | SetRequestArguments | EchoRequestArguments
}

type SetResponseArguments struct{}

type ResponseArguments interface {
	GetResponseArguments | SetResponseArguments
}

type Invocation[A Arguments] struct {
	Name      string `json:"name"`
	Arguments A      `json:"arguments"`
	CallId    string `json:"callId"`
}

func (i *Invocation[A]) UnmarshalJSON(b []byte) error {
	//fmt.Println(string(b))
	l := make([]interface{}, 3)
	err := json.Unmarshal(b, &l)
	//fmt.Println(l)
	if err != nil {
		return err
	}

	b1, err := json.Marshal(&l[1])
	if err != nil {
		return err
	}

	err = json.Unmarshal(b1, &i.Arguments)
	if err != nil {
		return err
	}

	i.Name = l[0].(string)
	i.CallId = l[2].(string)

	return nil
}

func (i Invocation[A]) MarshalJSON() ([]byte, error) {
	list := []interface{}{i.Name, i.Arguments, i.CallId}
	return json.Marshal(list)
}

type Request[T RequestArguments] struct {
	Using       []string        `json:"using"`
	MethodCalls []Invocation[T] `json:"methodCalls"`
	CreatedIds  map[Id]Id       `json:"createdIds"`
}

type Response struct {
	MethodResponses []Invocation[any] `json:"methodResponses"`
	CreatedIds      map[Id]Id         `json:"createdIds"`
	SessionState    string            `json:"sessionState"`
}

type Client struct {
	methodCalls []Invocation[RequestArguments]
	session     *Session
	token       string
	userId      string
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateClient(url, token string) (*Client, error) {
	//login := Login{
	//	Username: username,
	//	Password: appPassword,
	//}
	//b, err := json.Marshal(&login)
	//if err != nil {
	//	return nil, err
	//}
	//resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	//if err != nil {
	//	return nil, err
	//}

	//if resp.StatusCode != http.StatusOK {
	//	return nil, errors.New(resp.Status)
	//}
	//url_struct, err := neturl.Parse(url)
	//if err != nil {
	//	return nil, err
	//}
	//
	//q := url_struct.Query()
	//q.Set("access_token", token)
	//url_struct.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	//req.SetBasicAuth(username, appPassword)

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	session := Session{}
	err = json.NewDecoder(resp.Body).Decode(&session)
	if err != nil {
		return nil, err
	}

	d1, err := json.MarshalIndent(&session, "", "\t")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile("session.json", d1, 0644)
	if err != nil {
		return nil, err
	}

	return &Client{session: &session, token: token}, nil
}

type EchoResponse struct {
}

func (c Client) Do() (*Response, error) {
	using := make([]string, 0)
	for capability := range c.session.Capabilities {
		using = append(using, capability)
	}
	request := Request[RequestArguments]{
		Using:       using,
		MethodCalls: c.methodCalls,
	}

	defer func() {
		c.methodCalls = c.methodCalls[:0]
	}()

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("%s\n", string(body))

	//url_struct, err := neturl.Parse(c.session.ApiUrl)
	//if err != nil {
	//	return nil, err
	//}
	//
	//q := url_struct.Query()
	//q.Set("access_token", c.token)
	//url_struct.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPost, c.session.ApiUrl, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	//resp, err := http.Post(URL, "application/json", bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			//e := Error{}
			//err := json.NewDecoder(resp.Body).Decode(&e)
			b, err := ioutil.ReadAll(resp.Body)
			//println(string(b))
			if err != nil {
				return nil, err
			}
			return nil, errors.New(string(b))
		}
		return nil, errors.New(fmt.Sprintf("Bad status: %d", resp.StatusCode))
	}
	r := Response{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (c Client) Echo() {
	echo := Invocation[RequestArguments]{
		Name: MethodEcho,
		Arguments: map[string]string{
			"AccountId": "Hello, World",
		},
		CallId: fmt.Sprintf("%d", len(c.methodCalls)),
	}
	c.methodCalls = append(c.methodCalls, echo)
}

func (c *Client) GetContact() {
	echo := Invocation[RequestArguments]{
		Name: MethodGetContact,
		Arguments: GetRequestArguments{
			AccountId: "ucb4140a6",
		},
		CallId: fmt.Sprintf("%d", len(c.methodCalls)),
	}
	c.methodCalls = append(c.methodCalls, echo)
}

func (c *Client) SetContact(ifInState *string, create *map[Id]interface{}, update *map[Id]PatchObject, destroy *[]Id) {
	echo := Invocation[RequestArguments]{
		Name: MethodSetContact,
		Arguments: SetRequestArguments{
			AccountId: "ucb4140a6",
			IfInState: ifInState,
			Create:    create,
			Update:    update,
			Destroy:   destroy,
		},
		CallId: fmt.Sprintf("%d", len(c.methodCalls)),
	}
	c.methodCalls = append(c.methodCalls, echo)
}
