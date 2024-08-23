package main

import (
	"fmt"
	"log"
	"time"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"strings"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const bufferSize = 10
var funcBuffer []func()

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func(){
		<-sigs
		fmt.Println("Aborting...")
		for _, cmd := range funcBuffer {
			cmd()
		}
		os.Exit(1)
	}()

	//CreateDatabase()
	GetAll()
	TrackTime()
}

func GetActiveWindow() (string, error) {
	winId, err := exec.Command("xdotool", "getactivewindow").Output()

	if err != nil {
		return "", err
	}

	winClass, err := exec.Command("xprop", "-id", strings.Trim(string(winId), " "), "WM_CLASS").Output()

	if err != nil {
		return "", err
	}

	res := strings.Split(string(winClass), "\"")
	return res[len(res)-2], nil
}

func SaveBuffer(function func()) {
	funcBuffer = append(funcBuffer, function)

	if len(funcBuffer) >= bufferSize {
		for _, cmd := range funcBuffer {
			cmd()
		}
		funcBuffer = funcBuffer[:0]
	}
}

func SaveUsage(program string, spentTime float64) func() {
	return func() {
		db, err := sql.Open("sqlite3", "./timetracker.db")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		today := time.Now().Format("2006-01-02")

		_, err = db.Exec("insert or ignore into programs (program, day) values (?, ?)", program, today)
		if err != nil {
			panic(err)
		}

		_, err = db.Exec("update programs set time = time + ? where program = ? AND day = ?", spentTime, program, today)
		if err != nil {
			panic(err)
		}
	}
}

func GetAll() {
	db, err := sql.Open("sqlite3", "./timetracker.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	rows, err := db.Query("select * from programs")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var program string
		var day string
		var time float64
		err = rows.Scan(&program, &day, &time)
		if err != nil {
			panic(err)
		}

		fmt.Println(program, day, time)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}
}

func TrackTime() {
	var lastWindow string
	
	lastTime := time.Now().Unix()

	for true {
		currentWindow, err := GetActiveWindow()
		currentTime := time.Now().Unix()

		if err != nil {
			log.Fatal(err)
		}

		if len(currentWindow) > 0 {
			if len(lastWindow) > 0 {
				SaveBuffer(SaveUsage(lastWindow, float64(currentTime - lastTime)))
			}
			lastWindow = currentWindow
			lastTime = currentTime
		}

		time.Sleep(1 * time.Second)
	}
}

func CreateDatabase() {
	db, err := sql.Open("sqlite3", "./timetracker.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("create table programs ( program varchar(255) primary key, day date, time int default 0);")
	if err != nil {
		panic(err)
	}
}
