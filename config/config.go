package config

type ConfigStruct struct {
	Publickey  string
	Bottoken   string
	Appid      string
	Tasklambda string
}

var Config = ConfigStruct{
	Publickey:  "",
	Bottoken:   "",
	Appid:      "",
	Tasklambda: "",
}
