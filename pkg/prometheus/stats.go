package prometheus

type Stats struct {
	Data struct {
		Result []struct {
			Metric struct {
				AccessKey string `json:"access_key"`
				Dir       string `json:"dir"`
				Proto     string `json:"proto"`
				Service   string `json:"service"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
