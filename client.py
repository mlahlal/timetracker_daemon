import socket

def send_request(request):
    """Invia una richiesta al server e riceve la risposta."""
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client_socket.connect(('localhost', 65432))

    client_socket.send(request.encode('utf-8'))
    response = client_socket.recv(1024).decode('utf-8')

    client_socket.close()
    return response

# Esempi di richieste al server
print("Richiedendo l'utilizzo delle app...")
usage = send_request("get_usage")
print(f"Risposta del server: {usage}")

print("Reset dell'utilizzo delle app...")
response = send_request("reset_usage")
print(f"Risposta del server: {response}")

print("Richiedendo di nuovo l'utilizzo delle app...")
usage = send_request("get_usage")
print(f"Risposta del server: {usage}")

