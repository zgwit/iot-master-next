package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zgwit/iot-master-next/src/ctrler"
	"github.com/zgwit/iot-master-next/src/model"
	"github.com/zgwit/iot-master-next/src/plugin"

	"github.com/zgwit/iot-master-next/src/utils"
)

var (
	system_middlewares  = []model.MiddlewareType{}
	system_applications = []model.ApplicationType{}

	system_influx      plugin.Influx
	system_nsq_client  plugin.NsqClient
	system_nsq_server  plugin.NsqServer
	system_http_server plugin.HttpServer

	ctrler_Middleware   ctrler.MiddlewareCtrler
	ctrler_Application  ctrler.ApplicationCtrler
	ctrler_data         ctrler.DataCtrler
	ctrler_device       ctrler.DeviceCtrler
	ctrler_device_model ctrler.DeviceModelCtrler

	info_disabled         = true
	info_ready_middleware = false
)

func main() {

	go system_info()

	middleware_init()

	system_init()
	ctrler_init()

	application_init()

	system_loop()
}

func middleware_init() {

	if err := model.ReadMiddlewareConfig(&system_middlewares); err != nil {
		log.Println("system_middleware_init.error:", err)
		system_fail()
	}

	for index := range system_middlewares {

		if system_middlewares[index].Enable {
			system_middlewares[index].Run()
		}
	}

	model.MiddlewareReady(&system_middlewares)

	info_ready_middleware = true
}

func application_init() {

	if err := model.ReadApplicationConfig(&system_applications); err != nil {
		log.Println("system_application_init.error:", err)
		system_fail()
	}

	for index := range system_applications {

		if system_applications[index].Enable {
			system_applications[index].Run()
		}
	}
}

func system_init() {

	var influx_config plugin.InfluxConfig

	if err := utils.ReadFilekeyToObject("./config.txt", "influxdb", &influx_config); err != nil {
		log.Println("system_init.influxdb.config.error:", err)
		system_fail()
	}

	system_influx = plugin.NewInflux(influx_config)

	if result, err := system_influx.Ping(30); !result || err != nil {
		log.Println("system_init.influxdb.ping.error:", err)
		system_fail()
	}

	var nsq_config plugin.NsqConfig

	if err := utils.ReadFilekeyToObject("./config.txt", "nsq", &nsq_config); err != nil {
		log.Println("system_init.nsq.config.error:", err)
		system_fail()
	}

	system_nsq_client = plugin.NewNsqClient(nsq_config)
	system_nsq_server = plugin.NewNsqServer(nsq_config)

	if err := system_nsq_client.Connect(); err != nil {
		log.Println("system_init.nsq.client.connect.error:", err)
		system_fail()
	}

	var http_server_config plugin.HttpServerConfig

	if err := utils.ReadFilekeyToObject("./config.txt", "http_server", &http_server_config); err != nil {
		log.Println("system_init.http_server.config.error: ", err)
		return
	}

	system_http_server = plugin.NewHttpServer(http_server_config)

	V1 := system_http_server.Router.Group("/v1")
	{
		MIDDLEWARE := V1.Group("/middleware")
		{
			MIDDLEWARE.GET("/list", ctrler_Middleware.List)
			MIDDLEWARE.GET("/status", ctrler_Middleware.Status)
			MIDDLEWARE.GET("/control", ctrler_Middleware.Control)
			MIDDLEWARE.POST("/update", ctrler_Middleware.Update)
			MIDDLEWARE.GET("/log", ctrler_Middleware.Log)
		}

		APPLICATION := V1.Group("/application")
		{
			APPLICATION.GET("/list", ctrler_Application.List)
			APPLICATION.GET("/status", ctrler_Application.Status)
			APPLICATION.GET("/control", ctrler_Application.Control)
			APPLICATION.POST("/update", ctrler_Application.Update)
			APPLICATION.GET("/log", ctrler_Application.Log)
		}

		DEVICE_MODEL := V1.Group("/device_model")
		{
			DEVICE_MODEL.GET("/list", ctrler_device_model.List)
			DEVICE_MODEL.POST("/create", ctrler_device_model.Create)
			DEVICE_MODEL.GET("/find", ctrler_device_model.Find)
			DEVICE_MODEL.POST("/update", ctrler_device_model.Update)
			DEVICE_MODEL.DELETE("/delete", ctrler_device_model.Delete)
			DEVICE_MODEL.POST("/config", ctrler_device_model.Config)
		}

		DEVICE := V1.Group("/device")
		{
			DEVICE.GET("/list", ctrler_device.List)
			DEVICE.POST("/create", ctrler_device.Create)
			DEVICE.GET("/find", ctrler_device.Find)
			DEVICE.POST("/update", ctrler_device.Update)
			DEVICE.DELETE("/delete", ctrler_device.Delete)
			DEVICE.POST("/config", ctrler_device.Config)

			ATTRIBUTE := DEVICE.Group("/attribute")
			{
				ATTRIBUTE.GET("/realtime", ctrler_device.AttributeRealtime)
				ATTRIBUTE.GET("/history", ctrler_device.AttributeHistory)
			}

			EVENT := DEVICE.Group("/event")
			{
				EVENT.GET("/realtime", ctrler_device.EventRealtime)
				EVENT.GET("/history", ctrler_device.EventHistory)
			}
		}
	}

	go system_http_server.Running()
}

func ctrler_init() {

	if err := ctrler_Middleware.Init(&system_middlewares); err != nil {
		log.Println("ctrler_init.ctrler_Middleware.error:", err)
		system_fail()
	}

	if err := ctrler_Application.Init(&system_applications); err != nil {
		log.Println("ctrler_init.ctrler_Application.error:", err)
		system_fail()
	}

	if err := ctrler_data.Init(&system_influx, &system_nsq_server, &system_nsq_client); err != nil {
		log.Println("ctrler_init.ctrler_data.error:", err)
		system_fail()
	}

	if err := ctrler_device.Init(&system_influx, &system_nsq_client); err != nil {
		log.Println("ctrler_init.ctrler_device.error:", err)
		system_fail()
	}

	if err := ctrler_device_model.Init(&system_influx, &system_nsq_client); err != nil {
		log.Println("ctrler_init.ctrler_device_model.error:", err)
		system_fail()
	}
}

func system_info() {

	if info_disabled {
		log.Println("successfully started.")
	} else {
		log.Println("waiting for the system to start.")
	}

	for {
		time.Sleep(time.Second)

		if info_disabled {
			return
		}

		info, name_max_length := "", 16

		model.MiddlewareMaxNameLength(&system_middlewares, &name_max_length)
		model.ApplicationMaxNameLength(&system_applications, &name_max_length)

		info += fmt.Sprintf("%-*s%-*s%-*s%-*s \n%s \n",
			name_max_length+6, "PROGRAM NAME",
			name_max_length, "ENABLE",
			name_max_length, "STATUS",
			name_max_length, "RESTART COUNT",
			strings.Repeat("-", name_max_length*4+6),
		)

		info += model.MiddlewareInfo(&system_middlewares, name_max_length)
		info += model.ApplicationInfo(&system_applications, name_max_length)

		if !info_ready_middleware {
			info += fmt.Sprintf("\nwaiting for all the middleware to start%s\n", strings.Repeat(".", int(time.Now().Unix())%6+1))
		}

		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(info)
	}
}

func system_loop() {

	for {
		time.Sleep(time.Second)
	}
}

func system_fail() {

	info_disabled = true

	log.Println("failed to start, press any key to exit.")
	fmt.Scanf("\n")
	os.Exit(0)
}
