package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

/*-------------------------------------------------------------- */
/*--------------------------FONCTIONS--------------------------- */
/*-------------------------------------------------------------- */

func sendFile(conn *net.UDPConn, fileName string, addr *net.UDPAddr) {

	//On ouvre notre fichier
	var file, err = os.Open(fileName)
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
		fmt.Println("The file is", fi.Size(), "bytes long")

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
		segSize := 0      //taille du segment total (max 1024 octets)
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
				i := 0
				for i < 5 {
					header = header + "0"
					i++
				}
				//fmt.Println(header)
			} else if count < 100 {
				i := 0
				for i < 4 {
					header = header + "0"
					i++
				}
			} else if count < 1000 {
				i := 0
				for i < 3 {
					header = header + "0"
					i++
				}
			} else if count < 10000 {
				i := 0
				for i < 2 {
					header = header + "0"
					i++
				}
			} else if count < 100000 {
				i := 0
				for i < 1 {
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

			//Gestion de pertes de paquets
			//time.Sleep(1 * time.Second)
			handle(conn, header, addr, seg)

			//On reset nos buffers
			header = header[:0]
			seg = seg[:0]

		}
		//Fin de l'envoi : on envoit "FIN" au client
		_, err = conn.WriteToUDP([]byte("FIN"), addr)
	}
}

func handle(conn *net.UDPConn, header string, addr *net.UDPAddr, seg []byte) {
	//On créé un buffer vide capable de garder 5 segments windowSeg[]
	//windowSeg := make([]byte, 5120)
	//On append chaque segment a ce buffer
	//for timeout>0 :
	//si ACK reçu pas le bon
	//On prend le num de seq de l'ACK reçu (ex: 000001)
	//On parcours buffSeg
	//On prend les 6 premiers bytes du seg headerSeg
	//Si string.Contains(numACK, headerSeg)
	//On retransmet les paquets a partir de ce segment
	//Si ACK recu :
	//On vide buffSeg

	//Pour chaque paquet, on regarde le dernier ACK recu
	buffACK := make([]byte, 10)

	timeout, _ := time.ParseDuration("20ms")
	conn.SetReadDeadline(time.Now().Add(timeout))

	n, _, err := conn.ReadFromUDP(buffACK)
	//fmt.Println("contenu buffer", string(buffACK))

	//GESTION DES ERREURS (timeout ou autre erreur)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			//si on arrive là, c'est qu'on a pas reçu le premier ACK donc on retransmet la fenetre
			fmt.Println("Packet lost -> retransmition")
			_, _ = conn.WriteToUDP(seg, addr)

			//on vide le buffer
			buffACK = buffACK[:0]
			handle(conn, header, addr, seg)

		} else {
			fmt.Println(err)
		}
		return

	}

	if buffACK != nil {
		fmt.Println("Received message :", string(buffACK), n, "bytes")
		//Si c'est le numero du header actuel -> continue
		if strings.Contains(string(buffACK), header) {
			//Tout va bien
			buffACK = buffACK[:0]
			return

		} else {
			//Sinon on prend le numero de l'ACK et on retransmet
			//5 paquets a partir de ce numero d'ACK
			_, err = conn.WriteToUDP(seg, addr)
			buffACK = buffACK[:0]
			return

		}
	}

}

// La goroutine file gère les échanges client-serveur en lien avec le fichier en parallèle
func file(new_port int, addr *net.UDPAddr) {

	/*------OUVERTURE DE LA CONNEXION SUR LE NOUVEAU PORT------ */
	buffer := make([]byte, 1024)

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

	/*---------------RECUPERER LE NOM DU FICHIER---------------- */
	n, _, err := conn.ReadFromUDP(buffer)

	if err != nil {
		fmt.Println(err)
		return
	}

	buffer = buffer[:n-1]

	fileName := string(buffer)
	fmt.Println("Received message", n, "bytes:", fileName)

	/*--------------------ENVOYER LE FICHIER-------------------- */
	sendFile(conn, fileName, addr)
}

//La fonction add_conn fait le three-way handshake et attribue puis retourne le numéro de port pour l'envoi du fichier au client
//la fonction retourne le pointeur vers la connexion udp établie "conn"
func add_conn(addr *net.UDPAddr, buffer []byte, nbytes int, connection *net.UDPConn, new_port int) int {

	/*---------------------------------------------------------- */
	/*---------------------THREE-WAY HANDSHAKE------------------ */
	/*---------------------------------------------------------- */

	fmt.Println("-------------------------------------")
	fmt.Println("--------THREE-WAY HANDSHAKE----------")
	fmt.Println("-------------------------------------")

	//Si le message recu est un SYN
	if strings.Contains(string(buffer), "SYN") {

		fmt.Print("Received message ", nbytes, " bytes: ", string(buffer), "\n")
		fmt.Println("Sending SYN_ACK...")

		//Le serveur est pret : on envoie le SYN-ACK avec le nouveau port
		_, _ = connection.WriteToUDP([]byte("SYN-ACK"+strconv.Itoa(new_port)), addr)

		//On attend un ACK
		nbytes, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return -1
		}

		if strings.Contains(string(buffer), "ACK") {
			fmt.Println("Received message", nbytes, "bytes :", string(buffer))
			fmt.Println("Three-way handshake established !")
			fmt.Println("-------------------------------------")
			return new_port
		}

	}
	return -1

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
	fmt.Println("ResolveUDPAddr :", s)
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
	new_port := 1024 //on commence à 1024 et pas 1000 car les 1024 sont limités pour les utilisateurs normaux (non root par exemple)

	//Création d'une map de connections ouvertes : clé = @ip:port_init ; valeur = new_port
	current_conn := make(map[*net.UDPAddr]int)

	for {
		fmt.Println("current_conn : ", current_conn)

		//On lit le message recu et on le met dans le buffer
		nbytes, addr, err := connection.ReadFromUDP(buffer)
		fmt.Println("adresse addr", addr)
		if err != nil {
			fmt.Println(err)
			return

		} else if _, found := current_conn[addr]; !found {
			/* si l'adresse de connexion n'est pas dans la map :
			- on ajoute l'adresse à la map
			- on lance la connexion avec la fonction add_conn */

			current_conn[addr] = new_port //clé: addr ; valeur = current_conn[addr]

			new_port += 1 //on incrémente le new_port de 1 pour la prochaine connexion

			if new_port == 9999 { //si on arrive à la fin de la plage de port, on reboucle au début de cette plage
				new_port = 1024
			}
			new_udp_port := add_conn(addr, buffer, nbytes, connection, current_conn[addr])
			// on lancera la goroutine avec l'envoie du fichier
			go file(new_udp_port, addr)
		}
	}

}
