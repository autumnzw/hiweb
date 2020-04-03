package webcmd

//func Route() {
//	http.HandleFunc("/swagger/index", func(writer http.ResponseWriter, req *http.Request) {
//		t := template.New("swagger_index.html")
//		index, _ := t.Parse(swagger_index_templ)
//		index.Execute(writer, "")
//	})
//}

func CreateRoute(searchDir string, projectName string) error {
	config := &Config{
		ProjectName:        projectName,
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "./controllers",
		PropNamingStrategy: "",
	}
	return NewGen().Build(config)
}
