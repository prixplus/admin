package main

import (
	"database/sql"
	"fmt"
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var githubSecret = ""
var accounts = gin.Accounts{}
var pathToSh = ""

// We accept just command 'go help' for test
func MainHandler(c *gin.Context) {
	user := c.MustGet(gin.AuthUserKey).(string)
	cmd := c.PostForm("cmd")
	dir := c.PostForm("dir")

	values := gin.H{
		"title": "Prix!",
		"user":  user,
		"cmd":   cmd,
		"dir":   dir,
	}

	if len(cmd) > 0 && cmd == "go help" {
		out, err := exe_cmd(dir, cmd)
		if err != nil {
			values["err"] = err.Error()
		} else {
			values["out"] = out
		}
	}

	c.HTML(http.StatusOK, "main.tmpl", values)
}

func HookHandler(c *gin.Context) {
	signature := c.Request.Header.Get("X-Hub-Signature")

	if signature != githubSecret {
		log.Printf("Access dennied signature: %s from ip: %s", signature, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	out, err := exe_cmd("", "/bin/sh "+pathToSh)
	if err != nil {
		log.Println("Error executing sh:", err)
	} else {
		log.Println("Success executing sh:", out)
	}

	c.Data(200, "text", []byte("Sucess!"))
}

func main() {

	prixpath := os.Getenv("GOPATH") + "/src/github.com/prixplus/admin/"

	// If var $MODE is set to RELEASE,
	// than starts server in release mode
	mode := strings.ToLower(os.Getenv("MODE"))
	if mode == "release" && gin.IsDebugging() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Init DB connection
	db, err := InitDB()
	if err != nil {
		log.Fatal("Error initializing DB: ", err)
	}

	// Close DB when main returns
	defer db.Close()

	initVars(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.LoadHTMLGlob(prixpath + "templates/*")
	r.Static("/assets", "./assets")
	//r.StaticFS("/more_static", http.Dir("my_file_system"))
	r.StaticFile("/favicon.ico", "./favicon.ico")

	r.GET("/ping", func(c *gin.Context) {
		c.Data(200, "text", []byte("pong!"))
	})

	r.GET("/hook", HookHandler)

	auth := r.Group("/cmd")
	auth.Use(gin.BasicAuth(accounts))
	auth.GET("/", MainHandler)
	auth.POST("/", MainHandler)

	// Logging the mode server is starting
	log.Printf("Server starting in %s mode", gin.Mode())

	// Manners allows you to shut your Go webserver down gracefully, without dropping any requests
	manners.ListenAndServe(":8888", r)
}

func initVars(db *sql.DB) {

	// Getting accounts from database
	rows, err := db.Query("SELECT username, password FROM users")
	if err != nil {
		log.Fatal("Error getting users: ", err)
	}
	username := ""
	password := ""
	for rows.Next() {
		err = rows.Scan(&username, &password)
		if err != nil {
			log.Fatal("Error scaning user accounts: ", err)
		}
		accounts[username] = password
	}

	rows.Close()

	// Getting vars
	githubSecret = getVar(db, "github_secret")
	pathToSh = getVar(db, "path_to_sh")
}

func exe_cmd(dir string, cmd string) (string, error) {
	log.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	c := exec.Command(head, parts...)
	c.Dir = dir

	out, err := c.Output()

	log.Printf("err: %s\n out:%s\n", err, out)

	return string(out), err
}

func getVar(db *sql.DB, varName string) string {
	value := ""
	rows, err := db.Query(fmt.Sprintf("SELECT value FROM variables WHERE name='%s'", varName))
	if err != nil {
		log.Fatal("Error getting github secret: ", err)
	}
	if rows.Next() {
		err = rows.Scan(&value)
		if err != nil {
			log.Fatal("Error scaning github secret: ", err)
		}
		log.Printf("Var %s: %s", varName, value)
		return value
	} else {
		log.Fatal("Variable github secret not found!")
	}
	return value
}
