package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"html/template"
	"log"
	"net/http"
)

var DBFILE = "test.db"

type Page struct {
	Title string
	Body  []byte
}

func ce(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//Sections directly correlate to buckets
func loadPage(section string, title string) Page {
	db, err := bolt.Open(DBFILE, 0600, nil)
	ce(err)
	var body []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(section))
		body = append(body, b.Get([]byte(title))...)
		return nil
	})
	ce(err)
	db.Close()
	return Page{Title: title, Body: body}
}

//Pages are saved in a SECTION (bucket) by their TITLE (key) in a BODY (the key's value)
func savePage(section string, page Page) error {
	db, err := bolt.Open(DBFILE, 0600, nil)
	ce(err)
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := []byte(section)
		b, err := tx.CreateBucketIfNotExists(bucket)
		ce(err)
		err = b.Put([]byte(page.Title), page.Body)
		ce(err)
		return nil
	})
	ce(err)
	db.Close()
	return nil
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.URL.Path[6:]
	fmt.Println(body)
	fmt.Println(title)
	savePage("Main", Page{Title: title, Body: []byte(body)})
	http.Redirect(w, r, "/b/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[3:]
	p := loadPage("Main", title)
	t, err := template.ParseFiles("templates/edit.html")
	ce(err)
	fmt.Println("ran edit")
	t.Execute(w, p)
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path
	title = title[3:]
	p := loadPage("Main", title)
	t, err := template.ParseFiles("templates/browse.html")
	ce(err)
	fmt.Println("ran browse")
	t.Execute(w, p)
}

func setDB() {
	testpage := Page{Title: "testing", Body: []byte("this is a test bro")}
	err := savePage("Main", testpage)
	ce(err)
	err = savePage("Main", Page{"title", []byte("that was the title")})
	ce(err)
	testout := loadPage("Main", "testing")
	fmt.Println(testout.Title, string(testout.Body))
}

func main() {
	setDB()
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/b/", browseHandler)
	http.HandleFunc("/e/", editHandler)
	http.ListenAndServe(":8080", nil)
}
