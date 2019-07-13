// https://golang.org/doc/articles/wiki/
package main

import (
  "fmt"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "regexp"
  "strings"
)

type Page struct {
  Title string
  Body []byte
}

var templates = template.Must(template.ParseFiles("./tmpl/frontpage.html", "./tmpl/edit.html", "./tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
  filename := "./data/" + p.Title + ".txt"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
  filename := "./data/" + title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p* Page) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func frontpageHandler(w http.ResponseWriter, r *http.Request) {
  files, err := ioutil.ReadDir("./data/")
  if err != nil {
    log.Fatal(err)
  }
  var pages []*Page
  for _, f := range files {
    fmt.Println(f.Name())
    title := strings.Replace(f.Name(), ".txt", "", -1)
    p, err := loadPage(title)
    if err == nil {
      fmt.Println(p.Title)
      fmt.Println(string(p.Body))
    }
    pages = append(pages, p)
  }

  fmt.Println("--------------------")

  for _, page := range pages {
    fmt.Println(page.Title)
    fmt.Println(string(page.Body))
  }

  renderTemplate(w, "frontpage", nil)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

func main() {
  http.HandleFunc("/", frontpageHandler)
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  log.Fatal(http.ListenAndServe(":8080", nil))
}
