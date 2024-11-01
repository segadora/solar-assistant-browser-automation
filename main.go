package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type SolarAssistant struct {
	url     string
	user    string
	pass    string
	debug   bool
	browser *rod.Browser
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("unable to load .env file: %s", err)
	}

	path, _ := launcher.LookPath()
	u := launcher.New().Bin(path).MustLaunch()

	solarAssistant := &SolarAssistant{
		url:     os.Getenv("SOLAR_ASSISTANT_URL"),
		user:    os.Getenv("SOLAR_ASSISTANT_USER"),
		pass:    os.Getenv("SOLAR_ASSISTANT_PASS"),
		debug:   os.Getenv("SOLAR_ASSISTANT_DEBUG") == "1",
		browser: rod.New().ControlURL(u).MustConnect(),
	}

	defer solarAssistant.browser.MustClose()

	r := gin.New()
	r.Use(
		cors.Default(),
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/health", "/favicon.ico"}}),
		gin.Recovery(),
	)
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/work-mode-schedule", solarAssistant.updateWorkModeSchedule)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("unable to start server: %s", err)
	}
}

type WorkModeScheduleUpdate struct {
	From     string
	To       string
	Priority string
	Enabled  int
}

type WorkModeSchedule1Update struct {
	From     string `form:"schedule1[from]"`
	To       string `form:"schedule1[to]"`
	Priority string `form:"schedule1[priority]"`
	Enabled  int    `form:"schedule1[enabled]"`
}

type WorkModeSchedule2Update struct {
	From     string `form:"schedule2[from]"`
	To       string `form:"schedule2[to]"`
	Priority string `form:"schedule2[priority]"`
	Enabled  int    `form:"schedule2[enabled]"`
}

type WorkModeSchedule3Update struct {
	From     string `form:"schedule3[from]"`
	To       string `form:"schedule3[to]"`
	Priority string `form:"schedule3[priority]"`
	Enabled  int    `form:"schedule3[enabled]"`
}

type WorkModeSchedule4Update struct {
	From     string `form:"schedule4[from]"`
	To       string `form:"schedule4[to]"`
	Priority string `form:"schedule4[priority]"`
	Enabled  int    `form:"schedule4[enabled]"`
}

type UpdateScheduleRequest struct {
	Schedule1 WorkModeSchedule1Update `form:"schedule1"`
	Schedule2 WorkModeSchedule2Update `form:"schedule2"`
	Schedule3 WorkModeSchedule3Update `form:"schedule3"`
	Schedule4 WorkModeSchedule4Update `form:"schedule4"`
}

func (solarAssistant *SolarAssistant) updateWorkModeSchedule(c *gin.Context) {
	var request UpdateScheduleRequest
	if err := c.BindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad request",
		})
		return
	}

	page := solarAssistant.browser.MustPage(solarAssistant.url + "/power").MustWaitStable()

	if page.MustElement(".heading").MustText() == "Sign in" {
		log.Println("inputting login")

		page.MustElement("input#user_email").MustInput(solarAssistant.user)
		page.MustElement("input#user_password").MustInput(solarAssistant.pass).MustType(input.Enter)
		page.MustWaitStable()

		log.Println("signing in")

		page.MustNavigate(solarAssistant.url + "/power").MustWaitStable()
	}

	m := make(map[string]WorkModeScheduleUpdate)
	m["1"] = WorkModeScheduleUpdate{From: request.Schedule1.From, To: request.Schedule1.To, Priority: request.Schedule1.Priority, Enabled: request.Schedule1.Enabled}
	m["2"] = WorkModeScheduleUpdate{From: request.Schedule2.From, To: request.Schedule2.To, Priority: request.Schedule2.Priority, Enabled: request.Schedule2.Enabled}
	m["3"] = WorkModeScheduleUpdate{From: request.Schedule3.From, To: request.Schedule3.To, Priority: request.Schedule3.Priority, Enabled: request.Schedule3.Enabled}
	m["4"] = WorkModeScheduleUpdate{From: request.Schedule4.From, To: request.Schedule4.To, Priority: request.Schedule4.Priority, Enabled: request.Schedule4.Enabled}

	if !solarAssistant.updateWorkSchedule(page, m) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, solarAssistant.response(gin.H{
			"status":  "error",
			"message": "unable to find 'work mode schedule'",
		}, page))
		return
	}

	// workaround due to bug, possibly in ui
	time.Sleep(10 * time.Second)

	needUpdate := false
	for i := 1; i <= 4; i++ {
		row := strconv.Itoa(i)
		req := m[row]

		if req.Enabled != 0 {
			el := page.MustElement("input#work_mode_slot_" + row + "_enabled_check")
			currentBool := el.MustProperty("checked").Bool()
			needUpdate = (currentBool && req.Enabled == -1) || (!currentBool && req.Enabled == 1)

			if needUpdate {
				break
			}
		}
	}

	if needUpdate {
		log.Println("enabled is not set correctly, fixing")
		if !solarAssistant.updateWorkSchedule(page, m) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, solarAssistant.response(gin.H{
				"status":  "error",
				"message": "unable to find 'work mode schedule'",
			}, page))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (solarAssistant *SolarAssistant) updateWorkSchedule(page *rod.Page, m map[string]WorkModeScheduleUpdate) bool {
	containerEl := page.MustElement("#schedule")

	if !strings.Contains(containerEl.MustElement("h3").MustText(), "Work mode schedule") {
		return false
	}

	page.MustElement("a[phx-click=\"edit_schedule\"]").MustClick().MustWaitStable()

	for i := 1; i <= 4; i++ {
		row := strconv.Itoa(i)
		req := m[row]

		if req.From != "" {
			containerEl.MustElement("input#work_mode_slot_" + row + "_start").MustInput("").MustType(timeInput(req.From))
		}

		if req.To != "" {
			containerEl.MustElement("input#work_mode_slot_" + row + "_end").MustInput("").MustType(timeInput(req.To))
		}

		if req.Priority != "" {
			containerEl.MustElement("select#work_mode_slot_" + row + "_priority").MustSelect(req.Priority)
		}

		if req.Enabled != 0 {
			el := page.MustElement("input#work_mode_slot_" + row + "_enabled_check")
			currentBool := el.MustProperty("checked").Bool()
			if (currentBool && req.Enabled == -1) || (!currentBool && req.Enabled == 1) {
				el.MustClick().MustWaitStable()
			}
		}
	}

	containerEl.MustElement("button[type=submit]").MustClick().MustWaitStable()

	return true
}

func (solarAssistant *SolarAssistant) response(response gin.H, page *rod.Page) gin.H {
	if solarAssistant.debug {
		page.MustScreenshot(time.Now().String() + ".png")
	}

	return response
}

func timeInput(time string) (input.Key, input.Key, input.Key, input.Key) {
	switch time {
	case "00":
		return input.Digit0, input.Digit0, input.Digit0, input.Digit0
	case "01":
		return input.Digit0, input.Digit1, input.Digit0, input.Digit0
	case "02":
		return input.Digit0, input.Digit2, input.Digit0, input.Digit0
	case "03":
		return input.Digit0, input.Digit3, input.Digit0, input.Digit0
	case "04":
		return input.Digit0, input.Digit4, input.Digit0, input.Digit0
	case "05":
		return input.Digit0, input.Digit5, input.Digit0, input.Digit0
	case "06":
		return input.Digit0, input.Digit6, input.Digit0, input.Digit0
	case "07":
		return input.Digit0, input.Digit7, input.Digit0, input.Digit0
	case "08":
		return input.Digit0, input.Digit8, input.Digit0, input.Digit0
	case "09":
		return input.Digit0, input.Digit9, input.Digit0, input.Digit0
	case "10":
		return input.Digit1, input.Digit0, input.Digit0, input.Digit0
	case "11":
		return input.Digit1, input.Digit1, input.Digit0, input.Digit0
	case "12":
		return input.Digit1, input.Digit2, input.Digit0, input.Digit0
	case "13":
		return input.Digit1, input.Digit3, input.Digit0, input.Digit0
	case "14":
		return input.Digit1, input.Digit4, input.Digit0, input.Digit0
	case "15":
		return input.Digit1, input.Digit5, input.Digit0, input.Digit0
	case "16":
		return input.Digit1, input.Digit6, input.Digit0, input.Digit0
	case "17":
		return input.Digit1, input.Digit7, input.Digit0, input.Digit0
	case "18":
		return input.Digit1, input.Digit8, input.Digit0, input.Digit0
	case "19":
		return input.Digit1, input.Digit9, input.Digit0, input.Digit0
	case "20":
		return input.Digit2, input.Digit0, input.Digit0, input.Digit0
	case "21":
		return input.Digit2, input.Digit1, input.Digit0, input.Digit0
	case "22":
		return input.Digit2, input.Digit2, input.Digit0, input.Digit0
	case "23":
		return input.Digit2, input.Digit3, input.Digit0, input.Digit0
	}

	return input.Digit0, input.Digit0, input.Digit0, input.Digit0
}
