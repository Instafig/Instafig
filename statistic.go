package main

import (
	"log"
	"os"
	"time"

	"github.com/appwilldev/Instafig/conf"
	"github.com/gin-gonic/gin"
	influx "github.com/influxdata/influxdb/client/v2"
)

var (
	statisticCh       = make(chan *influx.Point, 100000)
	influxClient      influx.Client
	influxBatchPoints influx.BatchPoints
	microSecondUnit   = time.Microsecond / time.Nanosecond
)

func init() {
	if conf.StatisticEnable {
		var err error
		influxClient, err = influx.NewHTTPClient(
			influx.HTTPConfig{
				Addr:     conf.InfluxURL,
				Username: conf.InfluxUser,
				Password: conf.InfluxPassword,
			})
		if err != nil {
			log.Println("Failed to init influx client: ", err.Error())
			os.Exit(1)
		}

		influxBatchPoints, err = influx.NewBatchPoints(influx.BatchPointsConfig{
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
	now := time.Now()
	c.Next()
	clientData := getClientData(c)
	if clientData == nil {
		return
	}

	respTime := time.Now().Sub(now)
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
		"resp_time":  int(respTime / microSecondUnit),
	}

	p, _ := influx.NewPoint("client_request", tags, fields, time.Now())
	select {
	case statisticCh <- p:
	default:
		log.Println("failed to send influx point to channel")
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
	influxBatchPoints.AddPoint(p)
	if len(influxBatchPoints.Points()) < conf.InfluxBatchPointsCount {
		return
	}

	if err := influxClient.Write(influxBatchPoints); err != nil {
		log.Println("Failed to dump point to influxdb: ", err.Error())
	}
	influxBatchPoints, _ = influx.NewBatchPoints(
		influx.BatchPointsConfig{
			Database:  conf.InfluxDB,
			Precision: "s",
		})
}
