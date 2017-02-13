package main

import (
	"html/template"
	"net/http"
	"io/ioutil"
	"fmt"
	"regexp"
	//"errors"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/negroni"
	"github.com/gorilla/mux"
	"log"
)
/* Global Variables */
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("../../../../views/view.html",
	"../../../../views/edit.html",
	"../../../../views/login.html"))
var db sql.DB


type Page struct {
	Title string
	Body []byte
}

/* Function to save to database */
/*func (p *Page) save(db sql.DB) string {

	return err.Error()
	*//*filename := p.Title + ".txt"
	return ioutil.WriteFile(filename,p.Body,0600)*//*
}*/

func load(title string) (*Page, error){
	filename := title + ".txt"
	body,err := ioutil.ReadFile(filename)
	if err != nil{
		return nil, err
	}
	return &Page{title,body}, nil
}

/*func checkerror(err error){
	if err != nil{
		panic(err)
	}
}*/



func editHandler(w http.ResponseWriter, r *http.Request, title string, db sql.DB){
	//checkDBConnection(db)
	p,err := load(title)
	if err != nil{
		p = &Page{title,[]byte("")}
	}
	renderTemplate("edit",w,p)
}



func loginHandler(w http.ResponseWriter,r *http.Request){
	if r.FormValue("login") != "" {
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
	p := &Page{}
	renderTemplate("login",w,p)
}

func renderTemplate(tmpl string,w http.ResponseWriter,p *Page){
	err := templates.ExecuteTemplate(w,tmpl+".html",p)
	if err != nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
	}
}

/*func getTitle(w http.ResponseWriter, r *http.Request)(string,error){
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil{
		http.NotFound(w,r)
		return "",errors.New("No match found!")
	}
	return m[2],nil
}*/

func checkDBConnection(db sql.DB)(string){
	err := db.Ping()
	if err != nil{
		return ""
	}
	return ""
}

func makeHandler(fn func (w http.ResponseWriter, r *http.Request, str string, db sql.DB) ) http.HandlerFunc{
	return func(w http.ResponseWriter,r *http.Request){
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil{
			http.NotFound(w,r)
			return
		}
		fn(w,r,m[2],db)
	}
}

func main() {
	/* Discard default multiplexor  */
	router := mux.NewRouter()

	/* Open a database connection  */
	db, err := sql.Open("mysql", "testuser:batmanthor1_2@/test")
	if err != nil{
		println(err.Error())
	}
	defer db.Close()

	/* Check if database connection is alive */
	err = db.Ping()
	if err != nil{
		fmt.Println(err.Error())
	}

	router.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
		fmt.Fprintf(w,"Hi %s",r.FormValue("email"))
	})
	router.HandleFunc("/view/{name}", func (w http.ResponseWriter, r *http.Request){
		title := r.URL.Path[len("/view/"):]
		row:= db.QueryRow("SELECT title,description FROM wiki WHERE title=?", title)
		p := &Page{}
		err = row.Scan(&p.Title,&p.Body)
		if err != nil{
			http.Redirect(w, r, "/edit/"+title, http.StatusFound)
			return
		}

		renderTemplate("view",w,p)
	})
	router.HandleFunc("/edit/{name}", makeHandler(editHandler))
	router.HandleFunc("/save/{name}", func (w http.ResponseWriter, r *http.Request){
		title := r.URL.Path[len("/edit/"):]
		body := r.FormValue("body")
		p := &Page{title,[]byte(body)}
		row:= db.QueryRow("SELECT title,description FROM wiki WHERE title=?", p.Title)
		err := row.Scan(&p.Title,&p.Body)
		switch {
		case err == sql.ErrNoRows:
			_, err := db.Exec("INSERT INTO wiki(title,description) VALUES(?, ?)", p.Title, p.Body)
			if err != nil{
				panic(err)
			}
			http.Redirect(w,r,"/view/"+title,http.StatusFound)
		case err != nil:
			log.Fatal(err)
		default:
			_, err := db.Exec("UPDATE wiki SET description=?", p.Body)
			if err != nil{
				panic(err)
			}
			fmt.Println(p.Title)
			http.Redirect(w,r,"/view/"+title,http.StatusFound)
		}



	})
	router.HandleFunc("/login",loginHandler)

	/* Instantiate Negroni Classic */
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":4500")
}