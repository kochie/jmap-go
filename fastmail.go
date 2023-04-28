package jmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type FastmailClient struct {
	*Client
}

type LoginRequest struct {
	Username string `json:"username"`
}

type AuthenticationRequest struct {
	LoginId  string `json:"loginId"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Remember bool   `json:"remember"`
}

type AuthenticationResponse struct {
	Methods        []Method `json:"methods"`
	LoginID        string   `json:"loginId"`
	MayTrustDevice bool     `json:"mayTrustDevice"`
}

type Method struct {
	Type string `json:"type"`
}

type FastmailSession struct {
	Session
	AccessToken string `json:"accessToken"`
	UserId      string `json:"userId"`
}

func NewFastmailClient(username, password, otp string) (*FastmailClient, error) {
	payload, err := json.Marshal(&LoginRequest{
		Username: username,
	})
	resp, err := http.Post("https://www.fastmail.com/jmap/authenticate/", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	loginResponse := AuthenticationResponse{}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		return nil, err
	}

	idx := find("password", loginResponse.Methods)
	if idx < 0 {
		return nil, errors.New("no password method for authentication")
	}

	payload, err = json.Marshal(&AuthenticationRequest{
		LoginId:  loginResponse.LoginID,
		Type:     "password",
		Value:    password,
		Remember: true,
	})
	if err != nil {
		return nil, err
	}

	resp, err = http.Post("https://www.fastmail.com/jmap/authenticate/", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Password authentication failed: " + resp.Status)
	}

	fmt.Println("Hello")

	passwordResponse := AuthenticationResponse{}
	err = json.NewDecoder(resp.Body).Decode(&passwordResponse)
	if err != nil {
		return nil, err
	}

	idx = find("totp", passwordResponse.Methods)

	payload, err = json.Marshal(&AuthenticationRequest{
		LoginId:  loginResponse.LoginID,
		Type:     "totp",
		Value:    otp,
		Remember: true,
	})
	if err != nil {
		return nil, err
	}

	resp, err = http.Post("https://www.fastmail.com/jmap/authenticate/", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	fastmailSession := FastmailSession{}
	err = json.NewDecoder(resp.Body).Decode(&fastmailSession)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", fastmailSession)

	token := fastmailSession.AccessToken
	client, err := CreateClient("https://api.fastmail.com/.well-known/jmap", token)
	if err != nil {
		return nil, err
	}

	client.userId = fastmailSession.UserId

	fm := FastmailClient{
		client,
	}

	return &fm, nil
}

func find(s string, methods []Method) int {
	for i, method := range methods {
		if method.Type == s {
			return i
		}
	}
	return -1
}
