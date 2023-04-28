package jmap

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	neturl "net/url"
	"os"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt"
)

func (c Client) signURL(url string) (*string, error) {
	r, err := regexp.Compile("[?#].*")
	if err != nil {
		return nil, err
	}

	cleanUrl := string(r.ReplaceAll([]byte(url), []byte("")))
	//log.Println(cleanUrl)
	decodedValue, err := neturl.QueryUnescape(cleanUrl)

	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write([]byte(decodedValue))
	bs := h.Sum(nil)
	b64url := base64.RawURLEncoding.EncodeToString(bs)

	//log.Println(c.session.SigningId)
	//log.Println("b64: ", b64url)
	//log.Println("b64: ", "1TVssbz_bnXLWODR3ShbwOeCw3gcwffnWGTyhe5O7kQ")

	claims := &jwt.StandardClaims{
		Issuer:  c.session.SigningId,
		Subject: b64url,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//log.Println(c.session.SigningKey)
	sk, err := base64.RawStdEncoding.DecodeString(c.session.SigningKey)
	if err != nil {
		return nil, err
	}
	ss, err := token.SignedString(sk)
	if err != nil {
		return nil, err
	}

	return &ss, nil
}

func (c Client) DownloadFile(blobId Id, name, filetype string) error {

	var accountId string
	for capability, account := range c.session.PrimaryAccounts {
		if capability == "https://www.fastmail.com/dev/blob" {
			accountId = string(account)
		}
	}

	url := strings.Replace(c.session.DownloadUrl, "{accountId}", accountId, 1)
	url = strings.Replace(url, "{blobId}", string(blobId), 1)
	url = strings.Replace(url, "{name}", name, 1)
	url = strings.Replace(url, "{type}", neturl.QueryEscape(filetype), 1)

	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("u", "cb4140a6")
	u.RawQuery = q.Encode()
	//url += "&u=" + "cb4140a6"

	accessToken, err := c.signURL(u.String())
	if err != nil {
		return err
	}

	q.Set("access_token", *accessToken)
	u.RawQuery = q.Encode()
	//url += "&access_token=" + *accessToken

	//log.Println(url)
	//fmt.Println("we")

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	//req.Header.Add("Authorization", "Bearer "+c.token)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		log.Println(string(b))
		return errors.New(resp.Status)
	} else {
		t, err := mime.ExtensionsByType(filetype)
		//log.Println(t)
		if err != nil {
			return err
		}
		err = os.MkdirAll("pp/", 0750)
		if err != nil {
			return err
		}
		if len(t) < 1 {
			log.Println(filetype)
		}
		f, err := os.Create("pp/" + name + t[len(t)-1])
		if err != nil {
			return err
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		//os.WriteFile("", resp.Body, 0644)
	}

	return nil
}

func (c Client) UploadFile(accountId string, file io.Reader) (*File, error) {

	//fmt.Println(c.session.UploadUrl)
	url, err := neturl.Parse(strings.Replace(c.session.UploadUrl, "{accountId}", accountId, 1))
	if err != nil {
		return nil, err
	}
	accessToken, err := c.signURL(url.String())
	if err != nil {
		return nil, err
	}
	q := url.Query()
	q.Set("access_token", *accessToken)
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPost, url.String(), file)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.token)

	//log.Println(url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		log.Println(string(b))
		return nil, errors.New(resp.Status)
	}

	f := File{}
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return nil, err
	}

	return &f, nil
}
