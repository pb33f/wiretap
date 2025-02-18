package shared

func SetCORSHeaders(headers map[string][]string) {
	headers["Access-Control-Allow-Headers"] = []string{"*"}
	headers["Access-Control-Allow-Origin"] = []string{"*"}
	headers["Access-Control-Allow-Methods"] = []string{"OPTIONS,POST,GET,DELETE,PATCH,PUT"}
}
