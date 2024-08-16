import subprocess
import time
from collections import defaultdict

def get_active_window():
    """Restituisce il nome della finestra attiva usando xprop e xdotool."""
    try:
        win_id = subprocess.check_output(['xdotool', 'getactivewindow']).strip()
        win_class = subprocess.check_output(['xprop', '-id', win_id, 'WM_CLASS']).decode('utf-8')
        return win_class.split('"')[-2]
    except subprocess.CalledProcessError:
        return None

def track_time():
    usage_times = defaultdict(float)
    last_window = None
    last_time = time.time()

    try:
        while True:
            current_window = get_active_window()
            current_time = time.time()

            if current_window is not None:
                if last_window is not None:
                    usage_times[last_window] += current_time - last_time
                last_window = current_window
                last_time = current_time

            # sostituire questa stampa con l'inserimento dei dati in un database SQLite
            if int(current_time) % 60 == 0:
                print("Tempo di utilizzo delle applicazioni:")
                for app, duration in usage_times.items():
                    print(f"{app}: {duration:.2f} seconds")
                print("-" * 30)

            time.sleep(1)

    except KeyboardInterrupt:
        print("Monitoraggio interrotto.")
