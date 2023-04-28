package jmap

type Account struct {
	Name                string                 `json:"name"`
	IsPersonal          bool                   `json:"isPersonal"`
	IsReadOnly          bool                   `json:"isReadOnly"`
	AccountCapabilities map[string]interface{} `json:"accountCapabilities"`
}

type Session struct {
	Capabilities    map[string]interface{} `json:"capabilities"`
	Accounts        map[Id]Account         `json:"accounts"`
	PrimaryAccounts map[string]Id          `json:"primaryAccounts"`
	Username        string                 `json:"username"`
	ApiUrl          string                 `json:"apiUrl"`
	DownloadUrl     string                 `json:"downloadUrl"`
	UploadUrl       string                 `json:"uploadUrl"`
	EventSourceUrl  string                 `json:"eventSourceUrl"`
	State           string                 `json:"state"`
	SigningId       string                 `json:"signingId"`
	SigningKey      string                 `json:"signingKey"`
}
