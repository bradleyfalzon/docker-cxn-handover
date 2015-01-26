package main

import (
	"os"

	"github.com/bradleyfalzon/docker-cxn-handover/container"
	"github.com/droundy/goopt"
	"github.com/op/go-logging"
)

var cid = goopt.String([]string{"-i", "--id"}, "", "Container ID")
var verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{}, "be verbose", "")

var log = logging.MustGetLogger("example")

func init() {

	// Setup logger
	var format = logging.MustStringFormatter(
		"%{color}%{time:2006-02-01 15:04:05.000} %{shortfunc}:%{color:reset} %{message}",
	)

	backend := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), format)
	logging.SetBackend(backend)

	// Setup args

	goopt.Parse(nil)

}

func main() {

	c := container.NewDocker(*cid, container.OSExec)

	err := c.IsValid()
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = c.IsRunning()
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")

}
