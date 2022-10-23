package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var db *sql.DB

func initDB() {
	dbName := os.Getenv("MYSQL_DB")
	uname := os.Getenv("MYSQL_USER")
	pwd := os.Getenv("MYSQL_PASSWORD")
	var err error
	db, err = sql.Open("mysql", uname+":"+pwd+"@tcp(yellow_app_db:3306)/"+dbName)

	if err != nil {
		log.Fatal(err)
	}
	// db.SetMaxOpenConns(5)
	var version string

	err2 := db.QueryRow("SELECT VERSION()").Scan(&version)

	if err2 != nil {
		log.Fatal(err2)
	}

	log.Println(version)

	_, err3 := db.Exec("SET autocommit = 0")
	if err3 != nil {
		log.Fatal(err3)
	}
	_, err4 := db.Exec("set global innodb_print_all_deadlocks = 1")
	if err4 != nil {
		log.Fatal(err4)
	}
}

// func doRollback(ctx context.Context) {
// 	log.SetPrefix(fmt.Sprintf("%s ", ctx.Value("x-request-id")))
// 	_, err := db.ExecContext(ctx, "ROLLBACK")
// 	if err != nil {
// 		log.Println(err.Error())
// 	}
// }

func doDBOperation(ctx context.Context) error {
	log.SetPrefix(fmt.Sprintf("%s ", ctx.Value("x-request-id")))
	log.Println("Started....")
	// _, err := db.ExecContext(ctx, "START TRANSACTION")
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return err
	// }
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	createParent := fmt.Sprintf("INSERT INTO parent(`name`) VALUES ('%s')", randSeq(10))
	res, err := db.ExecContext(ctx, createParent)
	if err != nil {
		log.Println(err.Error())
		// doRollback(ctx)
		err = tx.Rollback()
		if err != nil {
			log.Println(err.Error())
		}
		return err
	}
	parentID, err := res.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		// doRollback(ctx)
		err = tx.Rollback()
		if err != nil {
			log.Println(err.Error())
		}
		return err
	}

	// err = tx.Commit()
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return err
	// }
	// log.Printf("Parent Inserted [%d]", parentID)

	// db.Exec("commit")
	// db.Exec("start transaction")

	// updateChild := fmt.Sprintf("update child set `name` = '%s' where `parent_id` = %v", randSeq(10), parentID)
	// _, err = db.Exec(updateChild)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return err
	// }

	// tx, err = db.BeginTx(ctx, nil)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return err
	// }

	var childID int64
	createChild := fmt.Sprintf("INSERT INTO child(`name`, `parent_id`) VALUES ('%s', %v)", randSeq(10), parentID)
	res, err = db.ExecContext(ctx, createChild)
	if err != nil {
		log.Printf("Parent [%d] Error [%s]\n", parentID, err.Error())
		// doRollback(ctx)
		err = tx.Rollback()
		if err != nil {
			log.Println(err.Error())
		}
		return err
	}
	childID, err = res.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		// doRollback(ctx)
		err = tx.Rollback()
		if err != nil {
			log.Println(err.Error())
		}
		return err
	}

	// _, err = db.ExecContext(ctx, "commit")
	// if err != nil {
	// 	log.Println(err.Error())
	// 	doRollback(ctx)
	// 	return err
	// }
	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Printf("Inserted Parent [%d] and Child [%d]\n", parentID, childID)
	return nil
}

func printCounts(tableName string) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&count)
	switch {
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Number of rows in %s are %d\n", tableName, count)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	failedCOunter := 0
	initDB()
	defer db.Close()
	rand.Seed(time.Now().UnixNano())
	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "x-request-id", uuid.New().String())
		err := doDBOperation(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			failedCOunter++
			return
		}
		w.Write([]byte(""))
	})

	r.Get("/counts", func(w http.ResponseWriter, r *http.Request) {
		printCounts("parent")
		printCounts("child")
		w.Write([]byte(fmt.Sprintf("%d", failedCOunter)))
	})
	http.ListenAndServe(":8080", r)
}
