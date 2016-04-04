package main

import (
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

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

	if len(cmd) > 0 {
		out, err := exe_cmd(dir, cmd)
		if err != nil {
			values["err"] = err.Error()
		} else {
			values["out"] = out
		}
	}

	c.HTML(http.StatusOK, "main.tmpl", values)
}

func main() {

	prixpath := os.Getenv("GOPATH") + "/src/github.com/prixplus/admin/"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

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

	rows, err := db.Query("SELECT username, password FROM users")
	if err != nil {
		log.Fatal("Error getting users: ", err)
	}

	accounts := gin.Accounts{}

	// Getting accounts from database
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

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(gin.BasicAuth(accounts))

	r.LoadHTMLGlob(prixpath + "templates/*")
	r.Static("/assets", "./assets")
	//r.StaticFS("/more_static", http.Dir("my_file_system"))
	r.StaticFile("/favicon.ico", "./favicon.ico")

	r.GET("/ping", func(c *gin.Context) {
		c.Data(200, "text", []byte("pong!"))
	})

	r.GET("/", MainHandler)
	r.POST("/", MainHandler)

	// Logging the mode server is starting
	log.Printf("Server starting in %s mode", gin.Mode())

	// Manners allows you to shut your Go webserver down gracefully, without dropping any requests
	manners.ListenAndServe(":"+port, r)
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
