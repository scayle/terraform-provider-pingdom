package api_types

type Check struct {
	Id                       int64         `json:"id"`
	Name                     string        `json:"name"`
	Resolution               int64         `json:"resolution"`
	SendNotificationWhenDown int64         `json:"sendnotificationwhendown"`
	NotifyAgainEvery         int64         `json:"notifyagainevery"`
	NotifyWhenBackup         bool          `json:"notifywhenbackup"`
	Created                  int64         `json:"created"`
	Type                     CheckTypes    `json:"type"`
	Hostname                 string        `json:"hostname"`
	Ipv6                     bool          `json:"ipv6"`
	ResponseTimeThreshold    int64         `json:"responsetime_threshold"`
	CustomMessage            string        `json:"custom_message"`
	IntegrationIds           []interface{} `json:"integrationids"`
	LastErrorTime            int64         `json:"lasterrortime"`
	LastTestTime             int64         `json:"lasttesttime"`
	LastResponseTime         int64         `json:"lastresponsetime"`
	LastDownStart            int64         `json:"lastdownstart"`
	LastDownEnd              int64         `json:"lastdownend"`
	Status                   string        `json:"status"`
	Tags                     []CheckTag    `json:"tags"`
	ProbeFilters             []string      `json:"probe_filters"`
	UserIDs                  []int64       `json:"userids"`
}

type CheckTypes struct {
	HTTP CheckHTTPOptions `json:"http"`
}

type CheckTag struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Count string `json:"count"`
}

type CheckHTTPOptions struct {
	VerifyCertificate bool              `json:"verify_certificate"`
	URL               string            `json:"url"`
	Encryption        bool              `json:"encryption"`
	Port              int64             `json:"port"`
	RequestHeaders    map[string]string `json:"requestheaders"`
	SSLDownDaysBefore int64             `json:"ssl_down_days_before"`
	Username          string            `json:"username"`
	Password          string            `json:"password"`
}
