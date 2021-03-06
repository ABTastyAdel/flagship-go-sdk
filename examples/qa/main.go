package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/abtasty/flagship-go-sdk"
	"github.com/abtasty/flagship-go-sdk/pkg/client"
	"github.com/abtasty/flagship-go-sdk/pkg/tracking"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var fsClients = make(map[string]*client.FlagshipClient)
var fsVisitors = make(map[string]*client.FlagshipVisitor)

// FsSession express infos saved in session
type FsSession struct {
	EnvID        string //"blvo2kijq6pg023l8edg"
	UseBucketing bool   //true
	VisitorID    string
}

func (s *FsSession) getClient() *client.FlagshipClient {
	fsC, _ := fsClients[s.EnvID]
	return fsC
}

func (s *FsSession) getVisitor() *client.FlagshipVisitor {
	fsV, _ := fsVisitors[s.EnvID]
	return fsV
}

// FSEnvInfo Binding env from JSON
type FSEnvInfo struct {
	EnvironmentID string `json:"environment_id" binding:"required"`
	Bucketing     bool   `json:"bucketing" binding:"required"`
}

// FSVisitorInfo Binding visitor from JSON
type FSVisitorInfo struct {
	VisitorID string                 `json:"visitor_id" binding:"required"`
	Context   map[string]interface{} `json:"context" binding:"required"`
}

// FSHitInfo Binding visitor from JSON
type FSHitInfo struct {
	HitType                string  `json:"t" binding:"required"`
	Action                 string  `json:"ea"`
	Value                  int64   `json:"ev"`
	TransactionID          string  `json:"tid"`
	TransactionAffiliation string  `json:"ta"`
	TransactionRevenue     float64 `json:"tr"`
	ItemName               string  `json:"in"`
	ItemQuantity           int     `json:"iq"`
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func initSession() gin.HandlerFunc {

	return func(c *gin.Context) {
		session := sessions.Default(c)
		fsSessInt := session.Get("fs_session")

		if fsSessInt == nil {
			envID := "blvo2kijq6pg023l8edg"
			fsClient, _ := flagship.Start(envID, client.WithBucketing())
			fsClients[envID] = fsClient
			fsSess := FsSession{
				EnvID:        envID,
				UseBucketing: true,
			}
			setFsSession(c, &fsSess)
		}

		printMemUsage()
	}
}

func getFsSession(c *gin.Context) *FsSession {
	session := sessions.Default(c)
	fsSessInt := session.Get("fs_session")
	fsSess := fsSessInt.(*FsSession)
	return fsSess
}

func setFsSession(c *gin.Context, fsS *FsSession) {
	session := sessions.Default(c)
	session.Set("fs_session", fsS)
	err := session.Save()

	if err != nil {
		log.Fatalf("Error on saved cookie : %v", err)
	}
}

func main() {
	router := gin.Default()
	store := cookie.NewStore([]byte("fs-go-sdk-demo-secret"))
	router.Use(sessions.Sessions("fs-go-sdk-demo", store))
	gob.Register(&FsSession{})

	router.Use(initSession())

	router.Static("/static", "examples/qa/assets")

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "static/")
	})

	router.GET("/currentEnv", func(c *gin.Context) {
		fsSession := getFsSession(c)
		c.JSON(http.StatusOK, gin.H{
			"env_id":    fsSession.EnvID,
			"bucketing": fsSession.UseBucketing,
		})
	})

	router.POST("/setEnv", func(c *gin.Context) {
		var json FSEnvInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var fsClient *client.FlagshipClient
		var err error
		if json.Bucketing {
			fsClient, err = flagship.Start(json.EnvironmentID, client.WithBucketing())
		} else {
			fsClient, err = flagship.Start(json.EnvironmentID)
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		fsClientExisting, _ := fsClients[fsSession.EnvID]
		if fsClientExisting != nil {
			fsClientExisting.Dispose()
			fsClientExisting = nil
		}
		fsClients[json.EnvironmentID] = fsClient
		setFsSession(c, &FsSession{
			EnvID:        json.EnvironmentID,
			UseBucketing: json.Bucketing,
		})

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.POST("/setVisitor", func(c *gin.Context) {
		var json FSVisitorInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		fsClient, _ := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor, err := fsClient.NewVisitor(json.VisitorID, json.Context)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = fsVisitor.SynchronizeModifications()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fsVisitors[fsSession.EnvID+"-"+json.VisitorID] = fsVisitor
		setFsSession(c, &FsSession{
			EnvID:        fsSession.EnvID,
			UseBucketing: fsSession.UseBucketing,
			VisitorID:    json.VisitorID,
		})

		flagInfos := fsVisitor.GetAllModifications()

		c.JSON(http.StatusOK, gin.H{"flags": flagInfos})
	})

	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/getFlag/:name", func(c *gin.Context) {
		var flag = c.Param("name")
		var flagType = c.Query("type")
		var activate = c.Query("activate")
		var defaultValue = c.Query("defaultValue")

		if flag == "" || flagType == "" || activate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("Missing flag name, type, activate or defaultValue")})
			return
		}

		fsSession := getFsSession(c)
		fsClient, _ := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor, _ := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		var value interface{}
		var err error
		shouldActivate, err := strconv.ParseBool(activate)

		if err == nil {
			switch flagType {
			case "bool":
				defVal, castErr := strconv.ParseBool(defaultValue)
				if castErr != nil {
					err = castErr
					break
				}

				value, err = fsVisitor.GetModificationBool(flag, defVal, shouldActivate)
				break
			case "number":
				defVal, castErr := strconv.ParseFloat(defaultValue, 64)
				if castErr != nil {
					err = castErr
					break
				}

				value, err = fsVisitor.GetModificationNumber(flag, defVal, shouldActivate)
				break
			case "string":
				value, err = fsVisitor.GetModificationString(flag, defaultValue, shouldActivate)
				break
			default:
				err = fmt.Errorf("Flag type %v not handled", flagType)
				break
			}
		}

		errString := ""
		status := http.StatusOK
		if err != nil {
			status = http.StatusBadRequest
			errString = err.Error()
		}

		c.JSON(status, gin.H{"value": value, "err": errString})
	})

	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.POST("/sendHit", func(c *gin.Context) {
		var json FSHitInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		fsClient, _ := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor, _ := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		hitType := json.HitType

		var hit tracking.HitInterface

		switch hitType {
		case "EVENT":
			hit = &tracking.EventHit{Action: json.Action, Value: json.Value}
		case "PAGE":
			hit = &tracking.PageHit{BaseHit: tracking.BaseHit{DocumentLocation: c.Request.URL.String()}}
		case "TRANSACTION":
			rand.Seed(time.Now().UnixNano())
			hit = &tracking.TransactionHit{TransactionID: json.TransactionID, Affiliation: json.TransactionAffiliation, Revenue: json.TransactionRevenue}
		case "ITEM":
			hit = &tracking.ItemHit{TransactionID: json.TransactionID, Name: json.ItemName, Quantity: json.ItemQuantity}
		}

		err := fsVisitor.SendHit(hit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "hitType": hitType})
	})

	router.Run(":8080")
}
