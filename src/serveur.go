package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"strconv"
	//"time"
)


func main() {
	/*---------------------------------------------------------- */
	/*-----------------------INITIALISATION--------------------- */
	/*---------------------------------------------------------- */

	//On récupère le port
	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Println("Usage : ./serveur <port>")
		return
	}
	if len(arguments) > 2 {
		fmt.Println("Usage : ./serveur <port>")
		return
	}
	PORT := ":" + arguments[1]

	//On récupère l'adresse de l'UDP endpoint (endpoint=IP:port)
	s, err := net.ResolveUDPAddr("udp4", PORT) 
	if err != nil {
		fmt.Println(err)
		return
	}
	//On créé un serveur UDP
	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()

	//On crée et initialise un objet buffer de type []byte et taille 1024
	buffer := make([]byte, 1024) 

	/*---------------------------------------------------------- */
	/*---------------------THREE-WAY HANDSHAKE------------------ */
	/*---------------------------------------------------------- */

	fmt.Println("-------------------------------------")
	fmt.Println("--------THREE-WAY HANDSHAKE----------")
	fmt.Println("-------------------------------------")

	//On lit le message recu et on le met dans le buffer
	nbytes, addr, err := connection.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	if strings.TrimSpace(string(buffer[0:nbytes])) == "STOP" {
		fmt.Println("Exiting UDP server!")
		return
	}
	//Si le message recu est un SYN
	if strings.Contains(string(buffer), "SYN"){
		
		fmt.Print("Received message ", nbytes," bytes: ", string(buffer),"\n")
		fmt.Println("Sending SYN_ACK...")
		
		//On créé un serveur UDP pour les messages avec le nouveau port
		new_port := 6667

		add, err := net.ResolveUDPAddr("udp4", (":" + strconv.Itoa(new_port))) 
		if err != nil {
			fmt.Println(err)
			return
		}

		conn, err := net.ListenUDP("udp4", add)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer conn.Close()

		//Le serveur est pret : on envoie le SYN-ACK avec le nouveau port
		_, err = connection.WriteToUDP([]byte("SYN-ACK"+strconv.Itoa(new_port)), addr)

		//On attend un ACK
		nbytes, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		if strings.Contains(string(buffer), "ACK"){
			fmt.Println("Received message", nbytes,"bytes: ", string(buffer))
			fmt.Println("Three-way handshake established !")
			fmt.Println("-------------------------------------")
		}
		/*---------------------------------------------------------- */
		/*---------------RECUPERER LE NOM DU FICHIER---------------- */
		/*---------------------------------------------------------- */
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		fileName := string(buffer)
		fmt.Println("Received message", n,"bytes:", fileName)
		
		/*---------------------------------------------------------- */
		/*--------------------ENVOYER LE FICHIER-------------------- */
		/*---------------------------------------------------------- */



	}
}
