#/bin/bash
./client1 127.0.0.2 "$1" "$2"&
./client1 127.0.0.3 "$1" "$2"&
./client1 127.0.0.4 "$1" "$2"&
wait
