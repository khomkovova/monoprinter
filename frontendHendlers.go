package main


import (
	"fmt"
	"io/ioutil"
	"net/http"
)
func Main(w http.ResponseWriter, r *http.Request){
	username := getUsernameFromCookie(r)
	if username == ""{
		fmt.Println("Bad cookie")
		//Load index.html
	}
	fmt.Println("Yor username =", username)
	//load user info page
}

func Signin(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("public/html/signin.html")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte(data))

}

func Signup(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("public/html/signup.html")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte(data))
}