package pkg

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type Data struct {
	Message string
}

func HandleRequest() {
	mux := http.NewServeMux() // router
	mux.HandleFunc("/", home)
	mux.HandleFunc("/ascii-art", ascii)
	log.Println("server is run at http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func home(w http.ResponseWriter, r *http.Request) { // wr + tab
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tp, err := template.ParseFiles("./html/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	err = tp.Execute(w, nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ascii(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	tmpl, err := template.ParseFiles("./html/index.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	text := r.FormValue("text")
	font := r.FormValue("font")
	// w.Write([]byte(text))
	// w.Write([]byte(font))
	result, status := asciidraw(font, text)
	switch status {
	case http.StatusBadRequest:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	case http.StatusInternalServerError:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	case http.StatusNotFound:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	res := Data{
		Message: result,
	}
	tmpl.Execute(w, res)
}

func asciidraw(style string, text string) (string, int) {
	checkName := fillUpAscii(style)
	if checkName != http.StatusOK {
		return "", checkName
	}
	str := strings.Split(text, "\r\n")
	removeNewline(&str)
	result := "\n"
	for _, s := range str {
		if s == "" {
			result += "\n"
			continue
		}
		temp, err := draw(s)
		if err != http.StatusOK {
			return "", err
		}
		result += temp
	}
	return result, http.StatusOK
}

var Ascii [128][8]string

func fillUpAscii(style string) int {
	if style != "standard" && style != "shadow" && style != "thinkertoy" {
		return http.StatusInternalServerError
	}
	// temp, err := os.Open("font/" + style + ".txt")
	temp, err := os.Open("./font/" + style + ".txt")
	if err != nil {
		return http.StatusInternalServerError
	}
	defer temp.Close()
	example := " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	scanner := bufio.NewScanner(temp)
	lineCount := 0
	for _, v := range example {
		scanner.Scan()
		for i := 0; i < 8; i++ {
			lineCount++
			scanner.Scan()
			Ascii[int(v)][i] = scanner.Text()
		}
	}
	if lineCount != 760 {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func removeNewline(input *[]string) {
	nowords := true
	for _, s := range *input {
		if len(s) > 0 {
			nowords = false
		}
	}
	if nowords {
		*input = (*input)[1:]
	}
}

func draw(s string) (string, int) {
	var result string
	for i := 0; i < 8; i++ {
		var temp string
		for _, v := range s {
			if v < 0 || v > 127 || Ascii[int(v)][0] == "" {
				return "", http.StatusBadRequest
			}
			temp += Ascii[int(v)][i]
		}
		result = result + temp + "\n"
	}
	return result, http.StatusOK
}

type ErrorData struct {
	StatusCode int
	StatusText string
}
