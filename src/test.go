package main

import (
	"fmt"
	"os"
)


func main() {
	fileName := "hey.txt"
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
			headerSize := 0 //segment ID sur 6 octets
			chunkSize := 1018 //chunk de données à envoyer

			//création des buffers
			header := make([]byte, headerSize)
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
				header = append(header, byte(count))
				if count < 10 {
					i:=0
					for (i < 5){
						header = append([]byte{0}, header...)
						i++
					}
					//fmt.Println(header)
				} else if count < 100 {
					i:=0
					for (i < 4){
						header = append([]byte{0}, header...)
						i++
					}
				} else if count < 1000 {
					i:=0
					for (i < 3){
						header = append([]byte{0}, header...)
						i++
					}
				} else if count < 10000 {
					i:=0
					for (i < 2){
						header = append([]byte{0}, header...)
						i++
					}
				} else if count < 100000 {
					i:=0
					for (i < 1){
						header = append([]byte{0}, header...)
						i++
					}
				}
				//On met le header dans le segment à envoyer
				seg = append(seg, header...)

				//On met le chunk de données dans le segment à envoyer
				seg = append(seg, chunk...)

				//On envoie le segment au client et on attend un ACK


				//On reset nos buffers
				header = header[:0]
				seg = seg[:0]
			}
				

			//Fin de l'envoi : on envoit "FIN" au client
		}

	return
}