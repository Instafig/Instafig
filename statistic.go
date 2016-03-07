package main

import (
	"log"
	//"os/exec"
	//"os"
	"time"
	//"path/filepath"

	"github.com/appwilldev/Instafig/conf"
	"github.com/gin-gonic/gin"
	influx "github.com/influxdata/influxdb/client/v2"

	"os"
)

var (
	statisticCh      = make(chan *influx.Point, 100000)
	influxClient     influx.Client
	influxBatchPoint influx.BatchPoints
	//logf *os.File

)

func init() {
	if conf.StatisticEnable {
		//cpath, err := os.Getwd()
		//if err != nil {
		//	log.Println("Failed to get current path:", err.Error())
		//	os.Exit(1)
		//}
		//logPath := filepath.Join(cpath, "statistic_log")
		//if err := exec.Command("mkdir", "-p", logPath).Run(); err != nil {
		//	log.Println("Failed to create statistic log dir in current path: ", err.Error())
		//	os.Exit(1)
		//}
		//
		//logf, err = os.OpenFile(filepath.Join(logPath, "statistic.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		//if err != nil {
		//	log.Println("Failed to create statistic log file: ", err.Error())
		//	os.Exit(1)
		//}
		var err error
		influxClient, err = influx.NewHTTPClient(influx.HTTPConfig{
			Addr:     conf.InfluxURL,
			Username: conf.InfluxUser,
			Password: conf.InfluxPassword,
		})
		if err != nil {
			log.Println("Failed to init influx client: ", err.Error())
			os.Exit(1)
		}

		influxBatchPoint, err = influx.NewBatchPoints(influx.BatchPointsConfig{
			Database:  conf.InfluxDB,
			Precision: "s",
		})
		if err != nil {
			log.Println("Failed to init influx batch points: ", err.Error())
			os.Exit(1)
		}
		go logStatisticTask()
	}
}

func StatisticHandler(c *gin.Context) {
	clientData := getClientData(c)
	if clientData == nil {
		return
	}

	tags := map[string]string{
		"node": conf.ClientAddr,
		"app":  clientData.AppKey,
	}

	fields := map[string]interface{}{
		"status":     getServiceStatus(c),
		"error_code": getServiceErrorCode(c),
		"ip":         clientData.Ip,
		"lang":       clientData.Lang,
		"os":         clientData.OSType,
		"osv":        clientData.OSVersion,
		"appv":       clientData.AppVersion,
		"deviceid":   clientData.DeviceId,
	}

	if len(statisticCh) > 100000/2 {
		log.Println("statistic chan is too much full: ", len(statisticCh))
	} else {
		p, _ := influx.NewPoint("client_request", tags, fields, time.Now())
		statisticCh <- p
	}
}

func logStatisticTask() {
	exitChan := make(chan struct{})
	for {
		go func() {
			defer func() { exitChan <- struct{}{} }()
			for {
				logStatistic(<-statisticCh)
			}
		}()

		<-exitChan
	}
}

func logStatistic(p *influx.Point) {
	influxBatchPoint.AddPoint(p)
	if len(influxBatchPoint.Points()) < 10 {
		return
	}

	if err := influxClient.Write(influxBatchPoint); err != nil {
		log.Println("Failed to dump point to influxdb: ", err.Error())
	} else {
		influxBatchPoint, _ = influx.NewBatchPoints(influx.BatchPointsConfig{
			Database:  conf.InfluxDB,
			Precision: "s",
		})
	}
}
