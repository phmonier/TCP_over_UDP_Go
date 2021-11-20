package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

// code from linode.com

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}
	PORT := ":" + arguments[1]

	s, err := net.ResolveUDPAddr("udp4", PORT) //vérifie que l'adresse et le port sont bien conformes au réseau (ici udp4)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()
	buffer := make([]byte, 1024) //on crée et initialise un objet buffer de type []byte et taille 1024
	rand.Seed(time.Now().Unix())

	for {
		//ON RECOIT ET AFFICHE LE MESSAGE DU CLIENT
		n, addr, err := connection.ReadFromUDP(buffer)
		fmt.Print("-> ", string(buffer[0:n-1]))

		if strings.TrimSpace(string(buffer[0:n])) == "STOP" {
			fmt.Println("Exiting UDP server!")
			return
		}

		//ON REPOND AU CLIENT
		/*data := []byte(strconv.Itoa(random(1, 1001))) //on renvoie un entier random sous form de string
		fmt.Printf("data: %s\n", string(data))
		_, err = connection.WriteToUDP(data, addr)*/
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		data := []byte(text + "\n")
		_, err = connection.WriteToUDP(data, addr)
		if strings.TrimSpace(string(data)) == "STOP" {
			fmt.Println("Exiting UDP serveur!")
			return
		}

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
