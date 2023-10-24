package main

import (
	"fmt"
	zgwit_ctrler "local/ctrler"
	zgwit_model "local/model"
	zgwit_plugin "local/plugin"
	zgwit_utils "local/utils"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	system_middlewares  = []zgwit_model.Middleware{}
	system_applications = []zgwit_model.Application{}

	system_influx      zgwit_plugin.Influx
	system_nsq_client  zgwit_plugin.NsqClient
	system_nsq_server  zgwit_plugin.NsqServer
	system_http_client zgwit_plugin.HttpClient
	system_http_server zgwit_plugin.HttpServer

	ctrler_Middleware  zgwit_ctrler.MiddlewareCtrler
	ctrler_Application zgwit_ctrler.ApplicationCtrler
	ctrler_data        zgwit_ctrler.DataCtrler
	ctrler_device      zgwit_ctrler.DeviceCtrler
	ctrler_model       zgwit_ctrler.ModelCtrler

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

	if err := zgwit_model.ReadMiddlewareConfig(&system_middlewares); err != nil {
		log.Println("system_middleware_init.error:", err)
		system_fail()
	}

	for index := range system_middlewares {

		if system_middlewares[index].Enable {
			system_middlewares[index].Run()
		}
	}

	zgwit_model.MiddlewareReady(&system_middlewares)

	info_ready_middleware = true
}

func application_init() {

	if err := zgwit_model.ReadApplicationConfig(&system_applications); err != nil {
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

	var influx_config zgwit_plugin.InfluxConfig

	if err := zgwit_utils.ReadFilekeyToObject("./config.txt", "influxdb", &influx_config); err != nil {
		log.Println("system_init.influxdb.config.error:", err)
		system_fail()
	}

	system_influx = zgwit_plugin.NewInflux(influx_config)

	if result, err := system_influx.Ping(); !result || err != nil {
		log.Println("system_init.influxdb.ping.error:", err)
		system_fail()
	}

	var nsq_config zgwit_plugin.NsqConfig

	if err := zgwit_utils.ReadFilekeyToObject("./config.txt", "nsq", &nsq_config); err != nil {
		log.Println("system_init.nsq.config.error:", err)
		system_fail()
	}

	system_nsq_client = zgwit_plugin.NewNsqClient(nsq_config)
	system_nsq_server = zgwit_plugin.NewNsqServer(nsq_config)

	if err := system_nsq_client.Connect(); err != nil {
		log.Println("system_init.nsq.client.connect.error:", err)
		system_fail()
	}

	var http_client_config zgwit_plugin.HttpClientConfig

	if err := zgwit_utils.ReadFilekeyToObject("./config.txt", "http_client", &http_client_config); err != nil {
		log.Println("system_init.http_server.config.error: ", err)
		return
	}

	system_http_client = zgwit_plugin.NewHttpClient(http_client_config)

	var http_server_config zgwit_plugin.HttpServerConfig

	if err := zgwit_utils.ReadFilekeyToObject("./config.txt", "http_server", &http_server_config); err != nil {
		log.Println("system_init.http_server.config.error: ", err)
		return
	}

	system_http_server = zgwit_plugin.NewHttpServer(http_server_config)

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
		}

		MODEL := V1.Group("/model")
		{
			MODEL.GET("/list", ctrler_model.List)
			MODEL.POST("/create", ctrler_model.Create)
			MODEL.POST("/update", ctrler_model.Update)
			MODEL.DELETE("/delete", ctrler_model.Delete)
			MODEL.POST("/config", ctrler_model.Config)
		}

		DEVICE := V1.Group("/device")
		{
			DEVICE.GET("/list", ctrler_device.List)
			// DEVICE.POST("/create", ctrler_device.Create)
			// DEVICE.POST("/update", ctrler_device.Update)
			// DEVICE.DELETE("/delete", ctrler_device.Delete)

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

	if err := ctrler_model.Init(&system_influx, &system_nsq_client); err != nil {
		log.Println("ctrler_init.ctrler_model.error:", err)
		system_fail()
	}
}

func system_info() {

	wait_timer := 1

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

		zgwit_model.MiddlewareMaxNameLength(&system_middlewares, &name_max_length)
		zgwit_model.ApplicationMaxNameLength(&system_applications, &name_max_length)

		info += fmt.Sprintf("%-*s%-*s%-*s%-*s \n%s \n",
			name_max_length+6, "PROGRAM NAME",
			name_max_length, "ENABLE",
			name_max_length, "STATUS",
			name_max_length, "RESTART COUNT",
			strings.Repeat("-", name_max_length*4+6),
		)

		info += zgwit_model.MiddlewareInfo(&system_middlewares, name_max_length)
		info += zgwit_model.ApplicationInfo(&system_applications, name_max_length)

		if !info_ready_middleware {

			if wait_timer < 6 {
				wait_timer++
			} else {
				wait_timer = 1
			}

			info += fmt.Sprintf("\nwaiting for all the middleware to start%s\n", strings.Repeat(".", wait_timer))
		}

		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(info)
	}
}

func system_loop() {

	for {

		response, err := system_http_client.POST("device/model/create", zgwit_plugin.HTTPQUERY{}, &zgwit_model.Model{
			Name:      "test",
			Id:        "model_test2",
			KeepAlive: 32,
		})
		if err != nil {
			fmt.Println("err", err)
		}

		fmt.Println(zgwit_utils.ToJson3(response))

		time.Sleep(time.Second)
	}
}

func system_fail() {

	info_disabled = true

	log.Println("failed to start, press any key to exit.")
	fmt.Scanf("\n")
	os.Exit(0)
}
