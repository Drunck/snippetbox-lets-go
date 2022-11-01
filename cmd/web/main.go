package main

import (
	"context"
	"flag"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"snippetbox.nurkuisa.net/internal/models"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

var ctx = context.Background()

func main() {
	var err error

	addr := flag.String("addr", "localhost:4000", "HTTP network address")
	dsn := flag.String("dsn", "postgres://web:789@localhost:5432/snippetbox", "Postgres data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := openDB(*dsn)

	if err != nil {
		errorLog.Fatalf("\nUnable to connection to database: %v\n", err)
	}
	defer conn.Close()

	infoLog.Println("Connected!")

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{
			DB: conn,
		},
		templateCache: templateCache,
		formDecoder:   formDecoder,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, dsn)

	if err != nil {
		return nil, err
	}
	if err = db.Ping(ctx); err != nil {
		return nil, err
	}
	return db, err
}
