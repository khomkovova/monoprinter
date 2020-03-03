package main


import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)
func Main(w http.ResponseWriter, r *http.Request){
	err, email := getEmailFromCookie(r)
	if err != nil{
		log.Println("Error: ", err)
		fmt.Println("Bad cookie")
		//Load index.html
	}
	fmt.Println("Yor username =", email)
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