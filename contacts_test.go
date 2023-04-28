package jmap

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetContact(t *testing.T) {
	a := assert.New(t)

	client := Client{}

	client.Echo()
	resp, err := client.Do()

	a.Nil(err, "Error should be empty")
	a.NotNil(resp, "Should be defined")
}

func TestClient_GetContact(t *testing.T) {
	a := assert.New(t)

	//appPassword := "4kqn3zxrgdzd9wge"
	username := "robert@kochie.io"
	password := "Pin0cchio2"
	//url := "https://api.fastmail.com/.well-known/jmap"

	client, err := NewFastmailClient(username, password, "246167")

	a.Nil(err)

	client.GetContact()
	resp, err := client.Do()

	a.Nil(err, "Error should be empty")
	a.NotNil(resp, "Should be defined")

	for _, response := range resp.MethodResponses {

		b, err := json.Marshal(response.Arguments)
		if err != nil {
			panic(err)
		}
		list := make([]Contact, 0)
		err = json.Unmarshal(b, &list)

		if err != nil {
			panic(err)
		}

		for _, l := range list {
			//fmt.Printf("%+v\n", l)
			if l.Avatar != nil {
				//fmt.Printf("%+v\n", l.Avatar)
				name := l.Avatar.Name
				if name == "" {
					name = l.FirstName + l.LastName
				}
				err = client.DownloadFile(l.Avatar.BlobId, name, l.Avatar.Type)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	fmt.Println(resp.MethodResponses[0])
}
