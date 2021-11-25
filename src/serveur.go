package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"strconv"
	"time"
)

/*-------------------------------------------------------------- */
/*--------------------------FONCTIONS--------------------------- */
/*-------------------------------------------------------------- */

func sendFile( conn *net.UDPConn, fileName string, addr *net.UDPAddr ) {

	//On ouvre notre fichier
	var file, err = os.Open(fileName) //fileName
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	//Si le fichier n'est pas vide 
	if file != nil {
		//On cherche la taille du fichier
		fi, err := file.Stat()
		if err != nil {
			fmt.Println(err)
			return		
		}
		bufferSize := fi.Size()
		fmt.Println("The file is",fi.Size() ,"bytes long")

		//On créé un buffer pour mettre le contenu du fichier dedans
		bufferSlice := make([]byte, bufferSize) 


		//On lit le fichier dans notre buffer
		nbytes, err := file.Read(bufferSlice)
		if err != nil {
			fmt.Println(err)
			return		
		}

		/*On va diviser le buffer en différents segments :*/

		//taille de nos buffers
		segSize := 0 //taille du segment total (max 1024 octets)
		chunkSize := 1018 //chunk de données à envoyer

		//création des buffers
		var header string
		seg := make([]byte, segSize)			
		chunk := make([]byte, chunkSize)

		//création d'un compteur pour compter nos paquets
		count := 0

		//création d'un décompte du nombre de bytes à envoyer
		countdown := nbytes

		//Début de la fragmentation :
		bytesToCopy := 0
	
		//tant que le fichier n'est pas vide
		for countdown > 0 {
			//On choisit la quantité de données à envoyer
			if countdown > chunkSize {
				bytesToCopy = chunkSize
			} else {
				bytesToCopy = countdown
			}
			//fmt.Println("---------------------------------------------------")
			
			//On copie les données dans le chunk
			copy(chunk, bufferSlice[:bytesToCopy])

			//Si on a moins de bytes à copier que la taille du chunk, on réduit la taille du chunk
			if bytesToCopy < chunkSize {
				chunk = chunk[:bytesToCopy]
			}
			//fmt.Println("chunk :", string(chunk))
			//fmt.Println("---------------------------------------------------")
			
			//On supprime le chunk du buffer
			bufferSlice = bufferSlice[bytesToCopy:]
			//fmt.Println("buffer:" , string(bufferSlice))
			
			//On décrémente le countdown
			countdown = countdown - bytesToCopy

			//On choisit un ID à mettre dans le header pour le segment en rajoutant les 0 nécessaires
			count = count + 1
			if count < 10 {
				i:=0
				for (i < 5){
					header = header + "0"
					i++
				}
				//fmt.Println(header)
			} else if count < 100 {
				i:=0
				for (i < 4){
					header = header + "0"
					i++
				}
			} else if count < 1000 {
				i:=0
				for (i < 3){
					header = header + "0"
					i++
				}
			} else if count < 10000 {
				i:=0
				for (i < 2){
					header = header + "0"
					i++
				}
			} else if count < 100000 {
				i:=0
				for (i < 1){
					header = header + "0"
					i++
				}
			}

			header = header + strconv.Itoa(count)

			//On met le header dans le segment à envoyer
			seg = append(seg, header...)

			//On met le chunk de données dans le segment à envoyer
			seg = append(seg, chunk...)
			//fmt.Println(string(seg))

			//On envoie le segment au client
			_, err = conn.WriteToUDP(seg, addr)

			//SLOW START
			time.Sleep(1 * time.Second)

			//On reset nos buffers
			header = header[:0]
			seg = seg[:0]

		}
		//Fin de l'envoi : on envoit "FIN" au client
		_, err = conn.WriteToUDP([]byte("FIN"), addr)

	}
}

/*-------------------------------------------------------------- */
/*-----------------------------MAIN----------------------------- */
/*-------------------------------------------------------------- */

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
		
		fmt.Print("Received message ", nbytes," bytes: ",string(buffer),"\n")
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
			fmt.Println("Received message", nbytes,"bytes :", string(buffer))
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

		buffer = buffer[:n-1]

		fileName := string(buffer)
		fmt.Println("Received message", n,"bytes:", fileName)
		
		/*---------------------------------------------------------- */
		/*--------------------ENVOYER LE FICHIER-------------------- */
		/*---------------------------------------------------------- */
		sendFile(conn, fileName, addr)

	}

}
