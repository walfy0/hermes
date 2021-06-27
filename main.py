import os
import sys
import time
import signal
import select
import socket

def gao():
    s = socket.socket()
    host = socket.gethostname()
    port = 12345
    s.connect((host,port))
    print(s.recv(1024))
    s.close()

if __name__ == '__main__':
    s = socket.socket()
    host = socket.gethostname()
    print(host)
    port = 12345
    s.bind((host, port))
    s.listen(5)
    while True:
        c,addr=s.accept()
        print('addr',addr)
        c.send(bytes('hello world','utf-8'))
        c.close()