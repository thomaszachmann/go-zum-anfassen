# Runbook: Disk voll
Wenn ein Node DiskPressure meldet, zuerst ungenutzte Images löschen:
crictl rmi --prune. Danach Logs unter /var/log prüfen und rotieren.
