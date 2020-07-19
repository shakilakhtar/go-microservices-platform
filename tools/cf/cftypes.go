package cf

type (
	cfCredentials struct {
		ID         int    `json:"ID"`
		BindingId  string `json:"binding_id"`
		Database   string `json:"database"`
		Dsn        string `json:"dsn"`
		Host       string `json:"host"`
		InstanceId string `json:"instance_id"`
		JdbcUri    string `json:"jdbc_uri"`
		Password   string `json:"password"`
		Port       string `json:"port"`
		Uri        string `json:"uri"`
		Username   string `json:"username"`
		Sslmode    string `json:"sslmode"`
	}

	CFSvcResponse struct {
		CFCredentials cfCredentials `json:"credentials"`
		Label         string        `json:"label"`
		Name          string        `json:"name"`
		Plan          string        `json:"plan"`
		Tags          []string      `json:"tags"`
	}

	CFDBService struct {
		Postgres []CFSvcResponse `json:"postgres"`
	}

)

