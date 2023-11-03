package main

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/KubeOperator/webkubectl/gotty/backend/localcommand"
	"github.com/KubeOperator/webkubectl/gotty/server"
	"github.com/KubeOperator/webkubectl/gotty/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "gotty"
	app.Version = Version
	os.Setenv("app.Version", app.Version)
	app.Usage = "Share your terminal as a web application"
	app.HideHelp = true

	appOptions := &server.Options{}
	if err := utils.ApplyDefaultValues(appOptions); err != nil {
		exit(err, 1)
	}
	backendOptions := &localcommand.Options{}
	if err := utils.ApplyDefaultValues(backendOptions); err != nil {
		exit(err, 1)
	}
	redisOptions := &server.RedisOptions{}
	if err := utils.ApplyDefaultValues(redisOptions); err != nil {
		exit(err, 1)
	}
	cliFlags, flagMappings, err := utils.GenerateFlags(appOptions, backendOptions, redisOptions)
	if err != nil {
		exit(err, 3)
	}

	app.Flags = append(
		cliFlags,
		&cli.StringFlag{
			Name:    "config",
			Value:   "~/.gotty",
			Usage:   "Config file path",
			EnvVars: []string{"GOTTY_CONFIG"},
		},
	)

	app.Action = func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			msg := "Error: No command given."
			cli.ShowAppHelp(c)
			exit(fmt.Errorf(msg), 1)
		}

		utils.ApplyFlags(cliFlags, flagMappings, c, appOptions, backendOptions, redisOptions)

		appOptions.EnableBasicAuth = c.IsSet("credential") || c.IsSet("credential-file")
		appOptions.EnableTLSClientAuth = c.IsSet("tls-ca-crt")

		err = appOptions.Validate()
		if err != nil {
			exit(err, 6)
		}
		err = redisOptions.Validate()
		if err != nil {
			exit(err, 6)
		}
		redisdb = &Redis{redis.NewClient(&redis.Options{
			Addr:     redisOptions.Addr,
			Password: redisOptions.Password,
			DB:       0,
		})}
		_, err := redisdb.Ping().Result()
		if err != nil {
			exit(err, 6)
		}

		args := c.Args().Slice()
		factory, err := localcommand.NewFactory(args[0], args[1:], backendOptions)
		if err != nil {
			exit(err, 3)
		}

		hostname, _ := os.Hostname()
		appOptions.TitleVariables = map[string]interface{}{
			"command":  args[0],
			"argv":     args[1:],
			"hostname": hostname,
		}

		srv, err := server.New(factory, appOptions, redisOptions)
		if err != nil {
			exit(err, 3)
		}

		ctx, cancel := context.WithCancel(context.Background())
		gCtx, gCancel := context.WithCancel(context.Background())
		log.Println("Welcome to use webkubectl.")
		log.Printf("GoTTY is starting with command: %s", strings.Join(c.Args().Slice(), " "))

		errs := make(chan error, 1)
		go func() {
			errs <- srv.Run(ctx, server.WithGracefulContext(gCtx))
		}()
		err = waitSignals(errs, cancel, gCancel)

		if err != nil && err != context.Canceled {
			fmt.Printf("Error: %s\n", err)
			exit(err, 8)
		}
		return nil
	}
	go func() {
		WatchWebshellLog()
	}()
	app.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func waitSignals(errs chan error, cancel context.CancelFunc, gracefulCancel context.CancelFunc) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-errs:
		return err

	case s := <-sigChan:
		switch s {
		case syscall.SIGINT:
			gracefulCancel()
			fmt.Println("C-C to force close")
			select {
			case err := <-errs:
				return err
			case <-sigChan:
				fmt.Println("Force closing...")
				cancel()
				return <-errs
			}
		default:
			cancel()
			return <-errs
		}
	}
}

func WatchWebshellLog() {
	// 创建文件/目录监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: webshell log watch panic,%s\n", err)
		}
		fmt.Println("webshell log watch end")
		watcher.Close()
	}()
	done := make(chan bool)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("webshell log watch panic")
				close(done)
			}
		}()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// 打印监听事件
				fmt.Printf("event: op: %s, name, %s\n", event.Op.String(), event.Name)
				// 如果是创建事件或者写入事件
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					err = redisdb.LPush("webshellLog", event.Name).Err()
					if err != nil {
						fmt.Printf("Error: %s\n", err)
					}
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	// 监听当前目录
	err = watcher.Add("/mnt/")
	if err != nil {
		fmt.Printf("Error: watcher.Add error,%s\n", err)
	}
	fmt.Println("webshell watcher start...")
	<-done
}

var redisdb *Redis

type Redis struct {
	*redis.Client
}
