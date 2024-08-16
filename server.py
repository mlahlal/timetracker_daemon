import socket
import time

class Server:
    app_usage = {
        "app1": 10,
        "app2": 20,
        "app3": 30
    }

    def __init__(self, address, port):
        self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.server_socket.bind((address, port))
        self.server_socket.listen(5)

    def start(self):
        print('Server in ascolto...')
        while True:
            client_socket, addr = self.server_socket.accept()
            print(f"Connessione accettata da {addr}")
            self.handle_client(client_socket)

    def handle_client(self, client_socket):
        while True:
            request = client_socket.recv(1024).decode('utf-8')
            if not request:
                break
            if request.startswith("get_usage"):
                response = str(self.app_usage)
            elif request.startswith("reset_usage"):
                for app in self.app_usage:
                    self.app_usage[app] = 0
                response = "Usage reset"
            else:
                response = "Unknown command"
    
            client_socket.send(response.encode('utf-8'))

        client_socket.close()
