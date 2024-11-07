package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type SolarAssistant struct {
	url   string
	user  string
	pass  string
	debug bool
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("unable to load .env file: %s", err)
	}

	solarAssistant := &SolarAssistant{
		url:   os.Getenv("SOLAR_ASSISTANT_URL"),
		user:  os.Getenv("SOLAR_ASSISTANT_USER"),
		pass:  os.Getenv("SOLAR_ASSISTANT_PASS"),
		debug: os.Getenv("SOLAR_ASSISTANT_DEBUG") == "1",
	}

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
	log.Println("receiving request to update schedule")

	var request UpdateScheduleRequest
	if err := c.BindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad request",
		})
		return
	}

	log.Println("go to power page")

	browser := rod.New().MustConnect()

	defer browser.MustClose()

	page := browser.MustPage(solarAssistant.url + "/power")

	page.MustScreenshot("test.png")
	page.MustWaitStable()

	if page.MustElement(".heading").MustText() == "Sign in" {
		log.Println("inputting login")

		page.MustElement("input#user_email").MustInput(solarAssistant.user)
		page.MustElement("input#user_password").MustInput(solarAssistant.pass).MustType(input.Enter)
		page.MustWaitStable()

		log.Println("signing in")

		page.MustNavigate(solarAssistant.url + "/power").MustWaitStable()
	}

	log.Println("generating map of schedules")

	m := make(map[string]WorkModeScheduleUpdate)
	if request.Schedule1.From != "" || request.Schedule1.To != "" || request.Schedule1.Priority != "" || request.Schedule1.Enabled != 0 {
		m["1"] = WorkModeScheduleUpdate{From: request.Schedule1.From, To: request.Schedule1.To, Priority: request.Schedule1.Priority, Enabled: request.Schedule1.Enabled}
	}
	if request.Schedule2.From != "" || request.Schedule2.To != "" || request.Schedule2.Priority != "" || request.Schedule2.Enabled != 0 {
		m["2"] = WorkModeScheduleUpdate{From: request.Schedule2.From, To: request.Schedule2.To, Priority: request.Schedule2.Priority, Enabled: request.Schedule2.Enabled}
	}
	if request.Schedule3.From != "" || request.Schedule3.To != "" || request.Schedule3.Priority != "" || request.Schedule3.Enabled != 0 {
		m["3"] = WorkModeScheduleUpdate{From: request.Schedule3.From, To: request.Schedule3.To, Priority: request.Schedule3.Priority, Enabled: request.Schedule3.Enabled}
	}
	if request.Schedule4.From != "" || request.Schedule4.To != "" || request.Schedule4.Priority != "" || request.Schedule4.Enabled != 0 {
		m["4"] = WorkModeScheduleUpdate{From: request.Schedule4.From, To: request.Schedule4.To, Priority: request.Schedule4.Priority, Enabled: request.Schedule4.Enabled}
	}

	if !solarAssistant.updateWorkSchedule(page, m) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, solarAssistant.response(gin.H{
			"status":  "error",
			"message": "unable to find 'work mode schedule'",
		}, page))
		return
	}

	log.Printf("waiting to see if form is correctly updated")

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

	time.Sleep(10 * time.Second)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (solarAssistant *SolarAssistant) updateWorkSchedule(page *rod.Page, m map[string]WorkModeScheduleUpdate) bool {
	containerEl := page.MustElement("#schedule")

	if !strings.Contains(containerEl.MustElement("h3").MustText(), "Work mode schedule") {
		return false
	}

	log.Println("pressing edit schedule")

	page.MustElement("a[phx-click=\"edit_schedule\"]").MustClick().MustWaitStable()

	for row, req := range m {
		log.Printf("updating schedule %s", row)

		if req.From != "" {
			log.Println("updating from")

			startEl := containerEl.MustElement("input#work_mode_slot_" + row + "_start")
			startEl.MustInput("")
			for _, c := range timeInput(req.From) {
				log.Printf("inputting key: %s", c.Info().Key)

				startEl.MustType(c)
			}
		}

		if req.To != "" {
			log.Println("updating to")

			endEl := containerEl.MustElement("input#work_mode_slot_" + row + "_end")
			endEl.MustInput("")
			for _, c := range timeInput(req.To) {
				log.Printf("inputting key: %s", c.Info().Key)

				endEl.MustType(c)
			}
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

		log.Printf("updated schedule %s", row)
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

func timeInput(time string) []input.Key {
	keys := make([]input.Key, 0)

	for _, char := range time {
		switch string(char) {
		case "0":
			keys = append(keys, input.Digit0)
		case "1":
			keys = append(keys, input.Digit1)
		case "2":
			keys = append(keys, input.Digit2)
		case "3":
			keys = append(keys, input.Digit3)
		case "4":
			keys = append(keys, input.Digit4)
		case "5":
			keys = append(keys, input.Digit5)
		case "6":
			keys = append(keys, input.Digit6)
		case "7":
			keys = append(keys, input.Digit7)
		case "8":
			keys = append(keys, input.Digit8)
		case "9":
			keys = append(keys, input.Digit9)
		}
	}

	return keys
}
