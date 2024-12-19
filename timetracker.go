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
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/joho/godotenv/autoload"
)

const bufferSize = 10
const syncLimit = 5
var funcBuffer []func()
var syncCounter int
var protocol string

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

	/*err := godotenv.Load()
	if err != nil {
		panic(err)
	}*/
	
	syncCounter = 0

	protocol = IsWaylandOrX11()
	_ = protocol

	//CreateDatabase()
	//GetAll()
	TrackTime()
	//SyncData()
}

func IsWaylandOrX11() (string) {
	//protocol, err := exec.Command("echo", "$XDG_SESSION_TYPE").Output()
	protocol = os.Getenv("XDG_SESSION_TYPE")

	return string(protocol)
}

func GetActiveWindow() (string, error) {
	if protocol == "wayland" {
		return GetActiveWindowKdeWl()
	} else {
		return GetActiveWindowX11()
	}
}

// NOT WORKING
func GetActiveWindowKdeWl() (string, error) {
	winId, err := exec.Command("kdotool", "getactivewindow").Output()

	if err != nil {
		return "", err
	}

	cmd := exec.Command("kdotool", "--debug", "getwindowname", string(winId))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	}

	return string(winId), nil
}

func GetActiveWindowX11() (string, error) {
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

	syncCounter++

	if len(funcBuffer) >= bufferSize {
		for _, cmd := range funcBuffer {
			cmd()
		}
		funcBuffer = funcBuffer[:0]
	}

	if syncCounter > syncLimit {
		SyncData()
		syncCounter = 0
	}
}

func SaveUsage(program string, spentTime float64) func() {
	return func() {
		db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
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

func SyncData() {
	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("select distinct program from programs")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	data := map[string]interface{}{}

	for rows.Next() {
		var program string

		err = rows.Scan(&program)
		if err != nil {
			panic(err)
		}

		ProgramRows, err := db.Query("select day, time from programs where program = ?", program)
		if err != nil {
			panic(err)
		}

		defer ProgramRows.Close()

		times := map[string]interface{}{}

		for ProgramRows.Next() {
			var day string
			var timeVar int

			err = ProgramRows.Scan(&day, &timeVar)
			if err != nil {
				panic(err)
			}

			dayFormatted, err := time.Parse(time.RFC3339, day)
			if err != nil {
				panic(err)
			}
			times[dayFormatted.Format("2006-01-02")] = timeVar
		}
		data[program] = times
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	
	//fmt.Println(bytes.NewBuffer(jsonData))

	req, err := http.NewRequest("POST", "http://localhost:3000/program", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	bodyString := string(body)
	//fmt.Println(bodyString)
}

func GetAll() {
	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
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

		//fmt.Println(program, day, time)
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
				//fmt.Println(lastWindow)
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

	_, err = db.Exec("create table programs ( program varchar(255), day date, time int default 0, primary key (program, day))")
	if err != nil {
		panic(err)
	}
}
