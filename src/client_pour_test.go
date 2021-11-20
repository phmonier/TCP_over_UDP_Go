package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a host:port string")
		return
	}
	ADRESSE := arguments[1]

	s, err := net.ResolveUDPAddr("udp4", ADRESSE) //vérifie que l'adresse et le port sont bien conformes au réseau (ici udp4)
	c, err := net.DialUDP("udp4", nil, s)         //nil : local adresse chosen ; Dial creates a client connection to the target s
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String()) //print l'adresse du client
	defer c.Close()

	for {

		//ENVOIE MESSAGE AU SERVEUR
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		data := []byte(text + "\n")
		_, err = c.Write(data)
		if strings.TrimSpace(string(data)) == "STOP" {
			fmt.Println("Exiting UDP client!")
			return
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		//RECEVOIR ET AFFICHER LA RÉPONSE DU SERVEUR
		buffer := make([]byte, 1024)       //on crée et initialise un objet buffer de type []byte et taille 1024
		n, _, err := c.ReadFromUDP(buffer) //on rempli le buffer avec le packet UDP
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Reply: %s\n", string(buffer[0:n]))
	}
}
