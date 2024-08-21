package main

import (
	"fmt"
	"time"
	"os/exec"
	"strings"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
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

func SaveUsage(program string, time float64) {
	db, err := sql.Open("sqlite3", "./timetracker.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("insert or ignore into programs (program) values (?)", program)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("update programs set time = time + ? where program = ?", time, program)
	if err != nil {
		panic(err)
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
		var time float64
		err = rows.Scan(&program, &time)
		if err != nil {
			panic(err)
		}

		fmt.Println(program, time)
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
			panic(err)
		}

		if len(currentWindow) > 0 {
			if len(lastWindow) > 0 {
				SaveUsage(lastWindow, float64(currentTime - lastTime))
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

	_, err := db.Exec("create table programs ( program varchar(255) primary key, time int default 0); )")
	if err != nil {
		panic(err)
	}
}
